package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/syncx"
)

var (
	// can't use one SingleFlight per conn, because multiple conns may share the same cache key.
	singleFlight = syncx.NewSingleFlight()
	stats        = cache.NewStat("store")
)

type Store interface {
	// 图上下文管理
	SaveGraphContext(ctx context.Context, graphContext GraphContext) error
	GetGraphContext(ctx context.Context, tenantId string, graphKey GraphKey, contextKey string) (GraphContext, error)
	DeleteGraphContext(ctx context.Context, tenantId string, graphKey GraphKey, contextKey string) error

	// 信息原子管理
	SaveInfoAtom(ctx context.Context, infoAtom InfoAtom) error
	GetInfoAtom(ctx context.Context, tenantId, id string) (InfoAtom, error)
	DeleteInfoAtom(ctx context.Context, tenantId, id string) error

	// 分布式锁
	Lock(ctx context.Context, resourceKey string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, resourceKey string) error
}

type BaseStore struct {
	KeyPrefix string
	// 图上下文的默认过期时间
	GraphContextTTL time.Duration
	// 信息原子的默认过期时间
	InfoAtomTTL           time.Duration
	CacheInterface        cache.Cache
	InfoAtomTypeQueryFunc InfoAtomTypeQueryFunc
}

func NewBaseStore(keyPrefix string, graphContextTTL time.Duration, infoAtomTTL time.Duration, infoAtomTypeQueryFunc InfoAtomTypeQueryFunc, cacheConf cache.CacheConf, opts ...cache.Option) (Store, error) {
	return &BaseStore{
		KeyPrefix:             keyPrefix,
		GraphContextTTL:       graphContextTTL,
		InfoAtomTTL:           infoAtomTTL,
		CacheInterface:        cache.New(cacheConf, singleFlight, stats, monc.ErrNotFound, opts...),
		InfoAtomTypeQueryFunc: infoAtomTypeQueryFunc,
	}, nil
}

func (s *BaseStore) GetKeyPrefix() string              { return s.KeyPrefix }
func (s *BaseStore) GetGraphContextTTL() time.Duration { return s.GraphContextTTL }
func (s *BaseStore) GetInfoAtomTTL() time.Duration     { return s.InfoAtomTTL }
func (s *BaseStore) GetCacheInterface() cache.Cache    { return s.CacheInterface }

// SaveGraphContext 保存图上下文
func (s *BaseStore) SaveGraphContext(ctx context.Context, graphContext GraphContext) error {
	if s.CacheInterface == nil {
		return ErrRedisClientNil
	}

	data, err := json.Marshal(graphContext)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMarshalFailed, err)
	}

	gc, ok := graphContext.(interface{ GetTenantId() string })
	if !ok {
		return ErrGraphContextImplement
	}

	tenantId := gc.GetTenantId()
	if tenantId == "" {
		return ErrTenantIdEmpty
	}

	gk, ok := graphContext.(interface{ GetGraphKey() GraphKey })
	if !ok {
		return ErrGraphKeyImplement
	}

	ck, ok := graphContext.(interface{ GetContextKey() string })
	if !ok {
		return ErrContextKeyImplement
	}

	key := s.graphContextKey(tenantId, gk.GetGraphKey(), ck.GetContextKey())
	err = s.CacheInterface.SetWithExpire(key, string(data), s.GraphContextTTL)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSaveGraphContext, err)
	}

	return nil
}

// GetGraphContext 获取图上下文
func (s *BaseStore) GetGraphContext(ctx context.Context, tenantId string, graphKey GraphKey, contextKey string) (GraphContext, error) {
	if s.CacheInterface == nil {
		return nil, ErrRedisClientNil
	}

	if tenantId == "" {
		return nil, ErrTenantIdEmpty
	}

	key := s.graphContextKey(tenantId, graphKey, contextKey)
	var result string
	err := s.CacheInterface.Get(key, &result)
	if err != nil {
		if err == monc.ErrNotFound {
			return nil, ErrItemNotFound
		}
		return nil, err
	}

	if result == "" {
		return nil, ErrItemNotFound
	}

	// 假设 BaseGraphContext 是 GraphContext 接口的主要或唯一具体实现
	var concreteGraphContext BaseGraphContext
	if err := json.Unmarshal([]byte(result), &concreteGraphContext); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnmarshalFailed, err)
	}

	return &concreteGraphContext, nil
}

// DeleteGraphContext 删除图上下文
func (s *BaseStore) DeleteGraphContext(ctx context.Context, tenantId string, graphKey GraphKey, contextKey string) error {
	if s.CacheInterface == nil {
		return ErrRedisClientNil
	}

	if tenantId == "" {
		return ErrTenantIdEmpty
	}

	key := s.graphContextKey(tenantId, graphKey, contextKey)
	err := s.CacheInterface.Del(key)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteGraphContext, err)
	}

	return nil
}

// SaveInfoAtom 保存信息原子
func (s *BaseStore) SaveInfoAtom(ctx context.Context, infoAtom InfoAtom) error {
	if s.CacheInterface == nil {
		return ErrRedisClientNil
	}

	// 获取信息原子的类型
	atomType := infoAtom.GetType()
	if atomType == nil {
		return ErrInfoAtomTypeEmpty
	}

	// 创建一个存储专用的结构体，不直接使用完整的InfoAtomType，而是用ID和版本来存储
	type storableInfoAtom struct {
		TenantId  string            `json:"tenantId"`
		ID        string            `json:"id"`
		Kind      InfoAtomTypeKey   `json:"kind"` // 使用InfoAtomTypeKey替代完整的InfoAtomType
		Source    string            `json:"source"`
		Timestamp int64             `json:"timestamp"`
		Labels    map[string]string `json:"labels"`
		Payload   map[string]any    `json:"payload"`
	}

	// 从原始InfoAtom创建可存储版本
	storableAtom := storableInfoAtom{
		TenantId: infoAtom.GetTenantId(),
		ID:       infoAtom.GetID(),
		Kind: InfoAtomTypeKey{
			ID: atomType.GetID(),
			// TenantId: atomType.GetTenantId(),
			// Name:      atomType.GetName(),    // 使用类型名称作为ID
			// Version:   atomType.GetVersion(), // 保存版本
		},
		Source:    infoAtom.GetSource(),
		Timestamp: infoAtom.GetTimestamp(),
		Labels:    infoAtom.GetLabels(),
		Payload:   infoAtom.GetPayload(),
	}

	// 序列化处理
	data, err := json.Marshal(storableAtom)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMarshalFailed, err)
	}

	if storableAtom.TenantId == "" {
		return ErrTenantIdEmpty
	}

	if storableAtom.ID == "" {
		return ErrIDEmpty
	}

	key := s.infoAtomKey(storableAtom.TenantId, storableAtom.ID)
	err = s.CacheInterface.SetWithExpire(key, string(data), s.InfoAtomTTL)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSaveInfoAtom, err)
	}

	return nil
}

// GetInfoAtom 获取信息原子
func (s *BaseStore) GetInfoAtom(ctx context.Context, tenantId, id string) (InfoAtom, error) {
	if s.CacheInterface == nil {
		return nil, ErrRedisClientNil
	}

	if tenantId == "" {
		return nil, ErrTenantIdEmpty
	}

	if id == "" {
		return nil, ErrIDEmpty
	}

	key := s.infoAtomKey(tenantId, id)
	var result string
	err := s.CacheInterface.Get(key, &result)
	if err != nil {
		if err == monc.ErrNotFound {
			return nil, ErrItemNotFound
		}
		return nil, err
	}

	if result == "" {
		return nil, ErrItemNotFound
	}

	// 辅助结构体用于反序列化，使用InfoAtomTypeKey存储类型ID和版本
	var tempAtomForUnmarshal struct {
		TenantId  string            `json:"tenantId"`
		ID        string            `json:"id"`
		Kind      InfoAtomTypeKey   `json:"kind"` // 使用InfoAtomTypeKey替代完整的InfoAtomType
		Source    string            `json:"source"`
		Timestamp int64             `json:"timestamp"`
		Labels    map[string]string `json:"labels"`
		Payload   map[string]any    `json:"payload"`
	}

	if err := json.Unmarshal([]byte(result), &tempAtomForUnmarshal); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnmarshalFailed, err)
	}

	// 使用InfoAtomTypeQueryFunc获取实际的InfoAtomType
	var atomType InfoAtomType
	if s.InfoAtomTypeQueryFunc != nil && (tempAtomForUnmarshal.Kind.ID != "") {
		atomType, err := s.InfoAtomTypeQueryFunc(ctx, tempAtomForUnmarshal.Kind)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrGetInfoAtomType, err)
		}
		if atomType == nil {
			// 如果找不到对应的类型，记录警告但继续处理
			// 这里可以根据实际需求决定是否要直接返回错误
			logx.WithContext(ctx).Infof("[Store] 无法找到信息原子类型 ID=%s", tempAtomForUnmarshal.Kind.ID)
		}
	}

	// 构建BasicInfoAtom实例
	finalAtom := &BasicInfoAtom{
		TenantId:  tempAtomForUnmarshal.TenantId,
		ID:        tempAtomForUnmarshal.ID,
		Type:      atomType, // 使用通过InfoAtomTypeQueryFunc获取的类型
		Source:    tempAtomForUnmarshal.Source,
		Timestamp: tempAtomForUnmarshal.Timestamp,
		Labels:    tempAtomForUnmarshal.Labels,
		Payload:   tempAtomForUnmarshal.Payload,
	}

	return finalAtom, nil
}

// DeleteInfoAtom 删除信息原子
func (s *BaseStore) DeleteInfoAtom(ctx context.Context, tenantId, id string) error {
	if s.CacheInterface == nil {
		return ErrRedisClientNil
	}

	if tenantId == "" {
		return ErrTenantIdEmpty
	}

	if id == "" {
		return ErrIDEmpty
	}

	key := s.infoAtomKey(tenantId, id)
	err := s.CacheInterface.Del(key)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteInfoAtom, err)
	}

	return nil
}

// 生成图上下文键
func (s *BaseStore) graphContextKey(tenantId string, graphKey GraphKey, contextKey string) string {
	return fmt.Sprintf("%sgraph:%s:%v:%s", s.KeyPrefix, tenantId, graphKey, contextKey)
}

// 生成信息原子键
func (s *BaseStore) infoAtomKey(tenantId, id string) string {
	return fmt.Sprintf("%sinfo:%s:%s", s.KeyPrefix, tenantId, id)
}

// Lock 尝试获取分布式锁
func (s *BaseStore) Lock(ctx context.Context, resourceKey string, ttl time.Duration) (bool, error) {
	if s.CacheInterface == nil {
		return false, ErrRedisClientNil
	}

	if resourceKey == "" {
		return false, ErrResourceKeyEmpty
	}

	lockKey := s.lockKey(resourceKey)
	// 使用当前时间戳作为锁的值，方便后续识别
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	// go-zero/core/stores/cache 不直接支持 SetNX 操作
	// 使用带有 TTL 的 SetWithExpire 函数代替，并检查是否成功设置
	success := false
	err := s.CacheInterface.SetWithExpire(lockKey, lockValue, ttl)
	if err == nil {
		success = true
	}

	return success, err
}

// Unlock 释放分布式锁
func (s *BaseStore) Unlock(ctx context.Context, resourceKey string) error {
	if s.CacheInterface == nil {
		return ErrRedisClientNil
	}

	if resourceKey == "" {
		return ErrResourceKeyEmpty
	}

	lockKey := s.lockKey(resourceKey)
	err := s.CacheInterface.Del(lockKey)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrReleaseLock, err)
	}

	return nil
}

// 生成锁的键
func (s *BaseStore) lockKey(resourceKey string) string {
	return fmt.Sprintf("%slock:%s", s.KeyPrefix, resourceKey)
}
