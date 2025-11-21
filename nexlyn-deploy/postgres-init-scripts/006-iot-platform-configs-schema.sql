-- ===============================
-- Nexlyn IoT平台配置管理数据库表结构
-- 版本: v1.0 - 多平台集成配置管理
-- 支持功能: Skylark、Webhook、Kafka等第三方平台集成配置
-- 数据架构: PostgreSQL存储元数据，Etcd存储完整配置（含敏感信息）
-- ===============================

-- 平台配置元数据表
-- 注: 完整的平台配置（包括认证信息等敏感数据）存储在Etcd中，key格式为 platform-{type}/{id}
-- 此表仅存储平台配置的基础元数据，用于快速查询、检索和权限验证
CREATE TABLE IF NOT EXISTS iot_platform_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),   -- 平台配置UUID（系统自动生成，用于构造Etcd key）
    name VARCHAR(100) NOT NULL,                      -- 平台配置显示名称
    type VARCHAR(20) NOT NULL,                       -- 平台类型（skylark, webhook, kafka等）
    description TEXT,                                -- 平台配置描述

    -- 状态控制
    enabled BOOLEAN DEFAULT TRUE,                    -- 是否启用该平台配置

    -- 多租户支持
    tenant_id UUID NOT NULL REFERENCES system_tenants(id) ON DELETE CASCADE, -- 所属租户

    -- 审计字段
    created_by UUID,                                 -- 创建者用户ID（UUID）
    updated_by UUID,                                 -- 最后修改者用户ID（UUID）
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE              -- 软删除时间
);

-- ===============================
-- 索引优化
-- ===============================

-- iot_platform_configs表索引
CREATE INDEX IF NOT EXISTS idx_iot_platform_configs_type ON iot_platform_configs(type);
CREATE INDEX IF NOT EXISTS idx_iot_platform_configs_enabled ON iot_platform_configs(enabled);
CREATE INDEX IF NOT EXISTS idx_iot_platform_configs_tenant_id ON iot_platform_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_iot_platform_configs_created_at ON iot_platform_configs(created_at);
CREATE INDEX IF NOT EXISTS idx_iot_platform_configs_deleted_at ON iot_platform_configs(deleted_at) WHERE deleted_at IS NOT NULL;

-- 复合索引（常用查询优化）
CREATE INDEX IF NOT EXISTS idx_iot_platform_configs_tenant_type ON iot_platform_configs(tenant_id, type);
CREATE INDEX IF NOT EXISTS idx_iot_platform_configs_tenant_enabled ON iot_platform_configs(tenant_id, enabled);

-- ===============================
-- 触发器 - 自动更新updated_at字段
-- ===============================

-- 更新iot_platform_configs的updated_at
CREATE OR REPLACE FUNCTION update_iot_platform_configs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_iot_platform_configs_updated_at
    BEFORE UPDATE ON iot_platform_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_iot_platform_configs_updated_at();

-- ===============================
-- 表注释
-- ===============================

COMMENT ON TABLE iot_platform_configs IS 'IoT平台配置元数据表，完整配置存储在Etcd中，key格式: platform-{type}/{id}';
COMMENT ON COLUMN iot_platform_configs.id IS '平台配置UUID（系统自动生成，用于构造Etcd key: platform-{type}/{id}）';
COMMENT ON COLUMN iot_platform_configs.type IS '平台类型：skylark=Skylark流程引擎，webhook=Webhook推送，kafka=Kafka消息队列';
COMMENT ON COLUMN iot_platform_configs.tenant_id IS '所属租户ID，平台配置在租户内隔离';
COMMENT ON COLUMN iot_platform_configs.enabled IS '是否启用该平台配置，禁用后将不再使用';

-- ===============================
-- 审计字段外键约束（在所有表创建后添加）
-- ===============================

-- iot_platform_configs 表审计字段外键
ALTER TABLE iot_platform_configs
    ADD CONSTRAINT fk_iot_platform_configs_created_by
    FOREIGN KEY (created_by) REFERENCES system_users(id) ON DELETE SET NULL;

ALTER TABLE iot_platform_configs
    ADD CONSTRAINT fk_iot_platform_configs_updated_by
    FOREIGN KEY (updated_by) REFERENCES system_users(id) ON DELETE SET NULL;
