package engine

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core/devices"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rezeropoint/go-skylark/v2/engine"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// IoT IoT引擎接口
type IoT interface {
	// 模板管理
	CreateTemplate(ctx context.Context, metadata core.TemplateMetadata, onlineConfig *core.OnlineDetectionConfig, businessConfig *core.DataProcessingConfig, controlConfig *core.DeviceControlConfig) (string, error)  // 创建模板，返回模板ID
	UpdateTemplate(ctx context.Context, id string, metadata core.TemplateMetadata, onlineConfig *core.OnlineDetectionConfig, businessConfig *core.DataProcessingConfig, controlConfig *core.DeviceControlConfig) error // 更新模板（onlineConfig/businessConfig/controlConfig为nil表示删除对应配置）
	DeleteTemplate(ctx context.Context, id string) error                                                                                                                                                               // 删除模板（级联删除所有配置）
	GetTemplate(ctx context.Context, id string) (*core.Template, error)                                                                                                                                                // 获取模板详情（包含元数据和配置）
	ListTemplates(ctx context.Context, query core.TemplateQuery) ([]*core.TemplateSummary, int64, error)                                                                                                               // 查询模板列表（返回摘要列表和总数）

	// 设备绑定管理
	BindDevice(ctx context.Context, metadata core.DeviceBindingMetadata) (string, error)                                        // 绑定设备，返回设备绑定记录ID
	GetDevice(ctx context.Context, deviceID string, tenantID string, orgIDs []string) (*core.DeviceBinding, error)              // 获取设备详情
	ListDevices(ctx context.Context, query core.DeviceBindingQuery) ([]*core.DeviceBindingSummary, int64, error)                // 查询设备列表（返回摘要列表和总数）
	ListUnboundDevices(ctx context.Context, tenantID string) ([]core.UnboundDevice, error)                                      // 获取未绑定设备列表（租户级查询，不限制组织）
	UpdateDevice(ctx context.Context, deviceID string, tenantID string, orgIDs []string, update core.DeviceBindingUpdate) error // 更新设备信息
	DeleteDevice(ctx context.Context, deviceID string, tenantID string, orgIDs []string) error                                  // 删除设备（物理删除，解绑即删除）

	// 标签管理
	CreateTag(ctx context.Context, metadata core.DeviceTagMetadata) (string, error)                   // 创建标签，返回标签ID
	GetTag(ctx context.Context, tagID string, tenantID string) (*core.DeviceTag, error)               // 获取标签详情
	ListTags(ctx context.Context, query core.DeviceTagQuery) ([]*core.DeviceTagSummary, int64, error) // 查询标签列表（返回摘要列表和总数）
	UpdateTag(ctx context.Context, tagID string, tenantID string, update core.DeviceTagUpdate) error  // 更新标签
	DeleteTag(ctx context.Context, tagID string, tenantID string) error                               // 删除标签

	// 元数据查询
	GetDeviceCategories() []core.DeviceCategoryInfo                      // 获取所有设备类别
	GetStandardFields(category core.DeviceCategory) []core.StandardField // 获取指定类别的标准字段列表

	// 平台配置管理
	CreatePlatform(ctx context.Context, metadata core.PlatformMetadata, config *core.PlatformConfig) (string, error)                                                   // 创建平台配置（元数据+配置），返回生成的UUID
	GetPlatform(ctx context.Context, platformType core.PlatformType, id string, tenantID string) (*core.Platform, error)                                               // 获取平台配置（包含完整配置）
	ListPlatforms(ctx context.Context, tenantID string, platformType *core.PlatformType, keyword string, page, pageSize int) ([]*core.Platform, int64, error)          // 列出平台配置（返回Platform，不包含Config）
	UpdatePlatform(ctx context.Context, platformType core.PlatformType, id string, tenantID string, metadata core.PlatformMetadata, config *core.PlatformConfig) error // 更新平台配置
	DeletePlatform(ctx context.Context, platformType core.PlatformType, id string, tenantID string) error                                                              // 删除平台配置

	// 时序数据查询
	QueryTimeSeries(ctx context.Context, query core.TimeSeriesQuery) (*core.TimeSeriesResult, error)            // 查询时序数据（支持原始数据和聚合）
	GetLatestValues(ctx context.Context, query core.LatestValuesQuery) ([]core.DeviceLatestValues, error)       // 获取设备字段的最新值
	GetDeviceStatistics(ctx context.Context, query core.DeviceStatisticsQuery) ([]core.DeviceStatistics, error) // 获取设备统计信息

	// AI Box算法任务管理（统一接口）
	CreateAIBoxTask(ctx context.Context, tenantID, deviceID string, orgIDs []string, task *devices.AIBoxAlgorithmTask) (string, error)      // 创建AI Box算法任务，返回TaskID
	UpdateAIBoxTask(ctx context.Context, tenantID, deviceID string, orgIDs []string, taskID string, task *devices.AIBoxAlgorithmTask) error // 更新AI Box算法任务
	DeleteAIBoxTask(ctx context.Context, tenantID, deviceID string, orgIDs []string, taskID string) error                                   // 删除AI Box算法任务
	ListAIBoxTasks(ctx context.Context, tenantID, deviceID string, orgIDs []string) ([]devices.AIBoxAlgorithmTask, error)                   // 查询AI Box算法任务列表
	GetAIBoxCapabilities(ctx context.Context, tenantID, deviceID string, orgIDs []string) (*devices.AIBoxCapabilities, error)               // 获取AI Box算法能力
	ControlAIBoxTask(ctx context.Context, tenantID, deviceID string, orgIDs []string, taskID string, controlCommand int) error              // 控制AI Box算法任务（启动/停止）
}

// New 创建IoT引擎实例
// dbConn: PostgreSQL数据库连接
// skylarkEngine: Skylark 流程引擎（可选，传入 nil 则禁用 Skylark 分发功能）
func New(ctx context.Context, config Config, dbConn sqlx.SqlConn, redisClient *redis.Redis, clickHouseConn driver.Conn, skylarkEngine engine.SkylarkEngine) (IoT, error) {
	return newIoTClient(ctx, config, dbConn, redisClient, clickHouseConn, skylarkEngine)
}
