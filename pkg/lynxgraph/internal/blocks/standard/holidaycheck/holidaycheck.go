package holidaycheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/zeromicro/go-zero/core/logx"
)

const HolidayCheckVersion = "v1"

// HolidayDate 假期日期条目（与 holiday-calendar API 响应格式一致）
type HolidayDate struct {
	Date   string `json:"date"`    // 日期，格式 "2006-01-02"
	Name   string `json:"name"`    // 名称
	NameCN string `json:"name_cn"` // 中文名称
	NameEN string `json:"name_en"` // 英文名称
	Type   string `json:"type"`    // public_holiday=法定假日, transfer_workday=调休工作日
}

// HolidayCalendar 假期日历（与 holiday-calendar API 响应格式一致）
type HolidayCalendar struct {
	Year   int           `json:"year"`   // 年份
	Region string        `json:"region"` // 地区代码，如 "CN"
	Dates  []HolidayDate `json:"dates"`  // 特殊日期列表
}

// cacheEntry 缓存条目
type cacheEntry struct {
	calendar  *HolidayCalendar
	fetchDate string // 获取数据时的日期（用于判断是否跨天）
}

// holidayCache 假期数据缓存
var (
	holidayCache sync.Map // map[cacheKey]cacheEntry
)

// Config 假期检查积木配置
type Config struct {
	DateSource         string `json:"dateSource"`         // atom=从信息原子时间戳提取日期（默认）, now=当前日期
	HolidayAPIURL      string `json:"holidayApiUrl"`      // 假期API地址，默认 https://unpkg.com/holiday-calendar@1.1.6/data/CN/
	TerminateIfHoliday bool   `json:"terminateIfHoliday"` // 可选，假期则终止流程（返回 false）
	Timezone           string `json:"timezone"`           // 时区，默认 Asia/Shanghai
}

// HolidayCheckBlock 假期检查积木
type HolidayCheckBlock struct {
	core.BaseLogicBlock
}

// NewHolidayCheckBlock 创建一个新的 HolidayCheckBlock 实例
func NewHolidayCheckBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &HolidayCheckBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeHolidayCheck,
			RawConfig: config,
		},
	}, nil
}

func (b *HolidayCheckBlock) GetID() string                { return b.ID }
func (b *HolidayCheckBlock) GetType() core.LogicBlockType { return b.Type }
func (b *HolidayCheckBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *HolidayCheckBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 设置默认值
	if cfg.DateSource == "" {
		cfg.DateSource = "atom"
	}
	if cfg.HolidayAPIURL == "" {
		cfg.HolidayAPIURL = "https://unpkg.com/holiday-calendar@1.1.6/data/CN/"
	}
	if cfg.Timezone == "" {
		cfg.Timezone = "Asia/Shanghai"
	}

	// 验证 dateSource
	if cfg.DateSource != "atom" && cfg.DateSource != "now" {
		return fmt.Errorf("dateSource 必须是 'atom' 或 'now'，当前值: %s", cfg.DateSource)
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行假期检查逻辑
func (b *HolidayCheckBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 1. 获取目标时间
	var targetTime time.Time
	if config.DateSource == "now" {
		targetTime = time.Now()
	} else {
		// 默认使用 atom 时间戳
		timestamp := execCtx.GetInfoAtom().GetTimestamp()
		if timestamp == 0 {
			targetTime = time.Now()
			logx.WithContext(ctx).Infof("[HolidayCheck] InfoAtom 时间戳为 0，使用当前时间")
		} else {
			targetTime = time.UnixMilli(timestamp)
		}
	}

	// 2. 转换到指定时区
	loc, err := time.LoadLocation(config.Timezone)
	if err != nil {
		logx.WithContext(ctx).Errorf("[HolidayCheck] 无效的时区: %s, 使用 UTC", config.Timezone)
		loc = time.UTC
	}
	targetTime = targetTime.In(loc)
	dateStr := targetTime.Format("2006-01-02")
	year := targetTime.Year()

	logx.WithContext(ctx).Infof("[HolidayCheck] 检查日期: %s, 星期: %s", dateStr, targetTime.Weekday())

	// 3. 获取假期数据
	calendar, err := fetchHolidayCalendar(ctx, config.HolidayAPIURL, year, config.Timezone)
	if err != nil {
		return false, fmt.Errorf("[HolidayCheck] 获取假期数据失败: %w", err)
	}

	// 4. 判断是否为工作日
	isWorkdayResult := isWorkday(dateStr, targetTime.Weekday(), calendar)

	logx.WithContext(ctx).Infof("[HolidayCheck] 日期 %s 是否为工作日: %v", dateStr, isWorkdayResult)

	// 5. 根据 terminateIfHoliday 决定是否终止流程
	if config.TerminateIfHoliday && !isWorkdayResult {
		logx.WithContext(ctx).Infof("[HolidayCheck] 日期 %s 是假期，终止流程", dateStr)
		return false, nil
	}

	return true, nil
}

// fetchHolidayCalendar 获取假期日历（带缓存，每天凌晨自动刷新）
func fetchHolidayCalendar(ctx context.Context, baseURL string, year int, timezone string) (*HolidayCalendar, error) {
	cacheKey := fmt.Sprintf("%s%d", baseURL, year)

	// 获取当前日期（用于判断是否跨天）
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	today := time.Now().In(loc).Format("2006-01-02")

	// 尝试从缓存获取
	if cached, ok := holidayCache.Load(cacheKey); ok {
		entry := cached.(cacheEntry)
		// 检查是否跨天（凌晨后首次访问需要刷新）
		if entry.fetchDate == today {
			return entry.calendar, nil
		}
		logx.Infof("[HolidayCheck] 缓存已过期（跨天），重新获取 %d 年假期数据", year)
	}

	// 构造 API URL
	url := baseURL
	if url[len(url)-1] != '/' {
		url += "/"
	}
	url += fmt.Sprintf("%d.json", year)

	logx.Infof("[HolidayCheck] 正在获取 %d 年假期数据: %s", year, url)

	// 发送 HTTP 请求
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析 JSON
	var calendar HolidayCalendar
	if err := json.Unmarshal(body, &calendar); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 存入缓存（记录获取日期）
	holidayCache.Store(cacheKey, cacheEntry{
		calendar:  &calendar,
		fetchDate: today,
	})

	logx.Infof("[HolidayCheck] 已加载 %d 年假期数据，共 %d 个特殊日期", year, len(calendar.Dates))

	return &calendar, nil
}

// isWorkday 判断指定日期是否为工作日
func isWorkday(dateStr string, weekday time.Weekday, calendar *HolidayCalendar) bool {
	// 查找该日期是否有特殊安排
	for _, holidayDate := range calendar.Dates {
		if holidayDate.Date == dateStr {
			switch holidayDate.Type {
			case "public_holiday":
				// 法定假日，不是工作日
				return false
			case "transfer_workday":
				// 调休工作日，是工作日
				return true
			}
		}
	}

	// 没有特殊安排，按周末判断
	return weekday != time.Saturday && weekday != time.Sunday
}

// GetHolidayCheckSpec 返回 HolidayCheckBlock 的规格
func GetHolidayCheckSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"HolidayCheck",
		HolidayCheckVersion,
		"检查指定日期是否为工作日，支持中国法定假期和调休判断",
		[]string{"input", "output"},
		map[string]any{
			"dateSource": map[string]any{
				"type":        "string",
				"title":       "日期来源",
				"description": "atom=从信息原子时间戳提取日期（默认），now=当前日期",
				"check":       "",
				"enum":        []string{"atom", "now"},
				"enumLabels":  []string{"信息原子时间戳", "当前时间"},
				"placeholder": "atom",
			},
			"holidayApiUrl": map[string]any{
				"type":        "string",
				"title":       "假期 API 地址",
				"description": "假期数据 API 基础 URL，默认使用 holiday-calendar",
				"check":       "",
				"placeholder": "https://unpkg.com/holiday-calendar@1.1.6/data/CN/",
			},
			"terminateIfHoliday": map[string]any{
				"type":        "boolean",
				"title":       "假期时终止",
				"description": "如果是非工作日，是否终止流程（返回 false）",
				"check":       "",
			},
			"timezone": map[string]any{
				"type":        "string",
				"title":       "时区",
				"description": "时区设置，默认 Asia/Shanghai",
				"check":       "",
				"placeholder": "Asia/Shanghai",
			},
		},
		[]core.ServiceType{}, // 不依赖外部服务
	)
}
