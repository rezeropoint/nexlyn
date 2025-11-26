-- HTTP 数据接收配置表
-- 用于存储 HTTP 数据接收配置的元数据
-- 完整配置存储在 Etcd 中，key 格式: http-receive/{configId}
--
-- 设计说明：
-- - HTTP 接收处理任意格式的 JSON 数据，不需要设备类别和标准字段的概念
-- - 只记录创建者（created_by），不记录更新者
-- - updated_at 由数据库自动更新

-- 创建 HTTP 数据接收配置表
CREATE TABLE IF NOT EXISTS iot_http_receive_configs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    enabled     BOOLEAN DEFAULT true,
    tenant_id   VARCHAR(50) NOT NULL,
    created_by  VARCHAR(50),
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_http_receive_tenant ON iot_http_receive_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_http_receive_enabled ON iot_http_receive_configs(enabled);
CREATE INDEX IF NOT EXISTS idx_http_receive_created_at ON iot_http_receive_configs(created_at DESC);

-- 添加表注释
COMMENT ON TABLE iot_http_receive_configs IS 'HTTP 数据接收配置表（处理任意格式的 JSON 数据）';
COMMENT ON COLUMN iot_http_receive_configs.id IS '配置ID（UUID）';
COMMENT ON COLUMN iot_http_receive_configs.name IS '配置名称';
COMMENT ON COLUMN iot_http_receive_configs.description IS '配置描述';
COMMENT ON COLUMN iot_http_receive_configs.enabled IS '是否启用';
COMMENT ON COLUMN iot_http_receive_configs.tenant_id IS '租户ID';
COMMENT ON COLUMN iot_http_receive_configs.created_by IS '创建者用户ID';
COMMENT ON COLUMN iot_http_receive_configs.created_at IS '创建时间';
COMMENT ON COLUMN iot_http_receive_configs.updated_at IS '更新时间（自动更新）';
