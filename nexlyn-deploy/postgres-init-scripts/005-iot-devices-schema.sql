-- ===============================
-- Nexlyn IoT设备管理系统数据库表结构
-- 版本: v1.0 - 传感器设备和模板管理
-- 支持功能: 多租户设备隔离、设备模板管理、在线状态追踪、设备分类
-- 数据架构: PostgreSQL存储元数据，Etcd存储模板配置，ClickHouse存储时序数据
-- ===============================

-- 设备模板元数据表
-- 注: 详细的模板配置（字段映射、在线检测规则、数据转换等）存储在Etcd中，key格式为 sensor-template-{id}
-- 此表仅存储模板的基础元数据，用于快速查询和展示
CREATE TABLE IF NOT EXISTS iot_sensor_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model VARCHAR(100) NOT NULL,                     -- 设备型号标识（业务标识，租户内唯一）
    name VARCHAR(200) NOT NULL,                      -- 模板显示名称
    category VARCHAR(50),                            -- 设备类别（environment_sensor, motion_sensor等）
    manufacturer VARCHAR(100),                       -- 设备厂商
    description TEXT,                                -- 模板描述
    version VARCHAR(20),                             -- 模板版本号

    -- 状态控制
    enabled BOOLEAN DEFAULT TRUE,                    -- 是否启用该模板

    -- Topic配置（用于管理和删除Etcd配置）
    online_topic_suffixes TEXT[] DEFAULT '{}',       -- 在线检测主题后缀列表（用于构建sensor-online/配置）
    business_topic_suffixes TEXT[] DEFAULT '{}',     -- 业务数据主题后缀列表（预留）
    control_topic_suffixes TEXT[] DEFAULT '{}',      -- 控制主题后缀列表（用于构建control-config/配置）

    -- 平台配置关联（用于删除约束检查）
    used_platform_ids TEXT[] DEFAULT '{}',           -- 使用的平台配置ID列表（从DispatchConfigs提取）

    -- 统计信息
    device_count INTEGER DEFAULT 0,                  -- 使用该模板的设备数量（冗余字段，提升查询性能）

    -- 多租户支持
    tenant_id UUID REFERENCES system_tenants(id) ON DELETE CASCADE,

    -- 审计字段
    created_by UUID,                                 -- 创建者用户ID（UUID）
    updated_by UUID,                                 -- 最后修改者用户ID（UUID）
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,             -- 软删除时间

    -- 唯一约束：model在租户内唯一
    CONSTRAINT uk_sensor_templates_model_tenant UNIQUE (model, tenant_id)
);

-- IoT设备绑定表（合并设备基础信息和绑定关系）
CREATE TABLE IF NOT EXISTS iot_device_bindings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(255) UNIQUE NOT NULL,          -- 设备唯一标识符（通常从MQTT消息中提取）
    device_name VARCHAR(255) NOT NULL,               -- 设备名称
    device_alias VARCHAR(200) DEFAULT '',            -- 设备别名（在组织内的显示名称）
    device_model VARCHAR(100) NOT NULL,              -- 设备型号（关联iot_sensor_templates.model）
    device_category VARCHAR(50),                     -- 设备类别（冗余字段，从模板同步）

    -- 设备基本信息
    description TEXT,                                -- 设备描述
    location VARCHAR(255),                           -- 安装位置
    installation_date DATE,                          -- 安装日期

    -- 设备状态
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'maintenance', 'error', 'decommissioned')),
    is_online BOOLEAN DEFAULT FALSE,                 -- 当前在线状态（从Redis同步）
    online_status_updated_at TIMESTAMP WITH TIME ZONE, -- 在线状态最后更新时间

    -- 时间追踪
    last_online_at TIMESTAMP WITH TIME ZONE,         -- 最后在线时间
    last_offline_at TIMESTAMP WITH TIME ZONE,        -- 最后离线时间
    last_data_at TIMESTAMP WITH TIME ZONE,           -- 最后接收数据时间

    -- 多租户和组织绑定
    tenant_id UUID NOT NULL REFERENCES system_tenants(id) ON DELETE CASCADE,        -- 所属租户
    org_id UUID NOT NULL REFERENCES system_organizations(id) ON DELETE CASCADE,     -- 所属组织

    -- 审计字段
    created_by UUID,                                 -- 创建者用户ID（UUID）
    updated_by UUID,                                 -- 最后修改者用户ID（UUID）
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- 外键约束（复合外键，确保租户隔离）
    CONSTRAINT fk_device_model FOREIGN KEY (device_model, tenant_id)
        REFERENCES iot_sensor_templates(model, tenant_id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE,

    -- 唯一约束：同一设备在同一租户的同一组织下只能绑定一次
    CONSTRAINT uk_device_tenant_org UNIQUE (device_id, tenant_id, org_id)
);

-- IoT设备标签定义表
CREATE TABLE IF NOT EXISTS iot_tag_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    label VARCHAR(100) NOT NULL,                    -- 标签名称
    description TEXT,                               -- 标签描述
    color VARCHAR(20),                              -- 标签颜色（用于前端展示）
    tenant_id UUID NOT NULL REFERENCES system_tenants(id) ON DELETE CASCADE, -- 租户隔离
    created_by UUID,                                -- 创建者用户ID（UUID）
    updated_by UUID,                                -- 最后修改者用户ID（UUID）
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,            -- 软删除时间

    -- 唯一约束：标签名称在租户内唯一
    CONSTRAINT uk_iot_tag_label_tenant UNIQUE (label, tenant_id)
);

-- IoT设备-标签关联表
CREATE TABLE IF NOT EXISTS iot_device_tag_relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(255) NOT NULL,                -- 设备ID（关联iot_device_bindings.device_id）
    tag_id UUID NOT NULL,                           -- 标签ID（关联iot_tag_definitions.id）
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- 外键约束
    CONSTRAINT fk_device_tag_device FOREIGN KEY (device_id)
        REFERENCES iot_device_bindings(device_id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    CONSTRAINT fk_device_tag_tag FOREIGN KEY (tag_id)
        REFERENCES iot_tag_definitions(id)
        ON DELETE CASCADE,

    -- 唯一约束：同一设备不能重复关联同一标签
    CONSTRAINT uk_device_tag UNIQUE (device_id, tag_id)
);

-- 传感器模板-标签关联表
CREATE TABLE IF NOT EXISTS iot_sensor_template_tag_relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL,                      -- 模板ID（关联iot_sensor_templates.id）
    tag_id UUID NOT NULL,                           -- 标签ID（关联iot_tag_definitions.id）
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- 外键约束
    CONSTRAINT fk_template_tag_template FOREIGN KEY (template_id)
        REFERENCES iot_sensor_templates(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_template_tag_tag FOREIGN KEY (tag_id)
        REFERENCES iot_tag_definitions(id)
        ON DELETE CASCADE,

    -- 唯一约束：同一模板不能重复关联同一标签
    CONSTRAINT uk_template_tag UNIQUE (template_id, tag_id)
);

-- ===============================
-- 索引优化
-- ===============================

-- iot_sensor_templates表索引
CREATE INDEX IF NOT EXISTS idx_iot_sensor_templates_model ON iot_sensor_templates(model);
CREATE INDEX IF NOT EXISTS idx_iot_sensor_templates_category ON iot_sensor_templates(category);
CREATE INDEX IF NOT EXISTS idx_iot_sensor_templates_enabled ON iot_sensor_templates(enabled);
CREATE INDEX IF NOT EXISTS idx_iot_sensor_templates_tenant_id ON iot_sensor_templates(tenant_id);
CREATE INDEX IF NOT EXISTS idx_iot_sensor_templates_created_at ON iot_sensor_templates(created_at);
CREATE INDEX IF NOT EXISTS idx_iot_sensor_templates_used_platform_ids ON iot_sensor_templates USING GIN (used_platform_ids);

-- iot_device_bindings表索引
CREATE INDEX IF NOT EXISTS idx_iot_bindings_device_id ON iot_device_bindings(device_id);
CREATE INDEX IF NOT EXISTS idx_iot_bindings_device_model ON iot_device_bindings(device_model);
CREATE INDEX IF NOT EXISTS idx_iot_bindings_device_category ON iot_device_bindings(device_category);
CREATE INDEX IF NOT EXISTS idx_iot_bindings_status ON iot_device_bindings(status);
CREATE INDEX IF NOT EXISTS idx_iot_bindings_is_online ON iot_device_bindings(is_online);
CREATE INDEX IF NOT EXISTS idx_iot_bindings_tenant_id ON iot_device_bindings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_iot_bindings_org_id ON iot_device_bindings(org_id);
CREATE INDEX IF NOT EXISTS idx_iot_bindings_last_online_at ON iot_device_bindings(last_online_at);
CREATE INDEX IF NOT EXISTS idx_iot_bindings_last_data_at ON iot_device_bindings(last_data_at);
CREATE INDEX IF NOT EXISTS idx_iot_bindings_created_at ON iot_device_bindings(created_at);

-- iot_tag_definitions表索引
CREATE INDEX IF NOT EXISTS idx_iot_tag_defs_label ON iot_tag_definitions(label);
CREATE INDEX IF NOT EXISTS idx_iot_tag_defs_tenant_id ON iot_tag_definitions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_iot_tag_defs_created_at ON iot_tag_definitions(created_at);

-- iot_device_tag_relations表索引
CREATE INDEX IF NOT EXISTS idx_iot_device_tag_device_id ON iot_device_tag_relations(device_id);
CREATE INDEX IF NOT EXISTS idx_iot_device_tag_tag_id ON iot_device_tag_relations(tag_id);
CREATE INDEX IF NOT EXISTS idx_iot_device_tag_created_at ON iot_device_tag_relations(created_at);

-- iot_sensor_template_tag_relations表索引
CREATE INDEX IF NOT EXISTS idx_iot_template_tag_template_id ON iot_sensor_template_tag_relations(template_id);
CREATE INDEX IF NOT EXISTS idx_iot_template_tag_tag_id ON iot_sensor_template_tag_relations(tag_id);
CREATE INDEX IF NOT EXISTS idx_iot_template_tag_created_at ON iot_sensor_template_tag_relations(created_at);

-- 复合索引（常用查询优化）
CREATE INDEX IF NOT EXISTS idx_iot_bindings_tenant_status ON iot_device_bindings(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_iot_bindings_tenant_model ON iot_device_bindings(tenant_id, device_model);
CREATE INDEX IF NOT EXISTS idx_iot_bindings_online_status ON iot_device_bindings(tenant_id, is_online, status);
CREATE INDEX IF NOT EXISTS idx_iot_bindings_tenant_org ON iot_device_bindings(tenant_id, org_id);
CREATE INDEX IF NOT EXISTS idx_iot_bindings_org_status ON iot_device_bindings(org_id, status);

-- 复合索引（用于时序数据查询中的设备ID解析和权限过滤）
-- 此索引优化查询: WHERE tenant_id = ? AND org_id = ANY(?)
CREATE INDEX IF NOT EXISTS idx_iot_bindings_tenant_org_device
    ON iot_device_bindings(tenant_id, org_id, device_id);

-- ===============================
-- 触发器 - 自动更新updated_at字段
-- ===============================

-- 更新iot_sensor_templates的updated_at
CREATE OR REPLACE FUNCTION update_iot_sensor_templates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_iot_sensor_templates_updated_at
    BEFORE UPDATE ON iot_sensor_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_iot_sensor_templates_updated_at();

-- 更新iot_device_bindings的updated_at
CREATE OR REPLACE FUNCTION update_iot_bindings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_iot_bindings_updated_at
    BEFORE UPDATE ON iot_device_bindings
    FOR EACH ROW
    EXECUTE FUNCTION update_iot_bindings_updated_at();

-- 更新iot_tag_definitions的updated_at
CREATE OR REPLACE FUNCTION update_iot_tag_definitions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_iot_tag_definitions_updated_at
    BEFORE UPDATE ON iot_tag_definitions
    FOR EACH ROW
    EXECUTE FUNCTION update_iot_tag_definitions_updated_at();

-- 更新iot_device_tag_relations的updated_at
CREATE OR REPLACE FUNCTION update_iot_device_tag_relations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_iot_device_tag_relations_updated_at
    BEFORE UPDATE ON iot_device_tag_relations
    FOR EACH ROW
    EXECUTE FUNCTION update_iot_device_tag_relations_updated_at();

-- 更新iot_sensor_template_tag_relations的updated_at
CREATE OR REPLACE FUNCTION update_iot_sensor_template_tag_relations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_iot_sensor_template_tag_relations_updated_at
    BEFORE UPDATE ON iot_sensor_template_tag_relations
    FOR EACH ROW
    EXECUTE FUNCTION update_iot_sensor_template_tag_relations_updated_at();

-- ===============================
-- 触发器 - 同步设备模板的device_count
-- ===============================

-- 创建或更新设备时，增加模板的device_count
CREATE OR REPLACE FUNCTION increment_template_device_count()
RETURNS TRIGGER AS $$
BEGIN
    -- 如果是INSERT或UPDATE时device_model改变
    IF (TG_OP = 'INSERT') THEN
        UPDATE iot_sensor_templates
        SET device_count = device_count + 1
        WHERE model = NEW.device_model;
    ELSIF (TG_OP = 'UPDATE' AND OLD.device_model != NEW.device_model) THEN
        -- 旧模板计数减1
        UPDATE iot_sensor_templates
        SET device_count = device_count - 1
        WHERE model = OLD.device_model;
        -- 新模板计数加1
        UPDATE iot_sensor_templates
        SET device_count = device_count + 1
        WHERE model = NEW.device_model;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_increment_device_count
    AFTER INSERT OR UPDATE OF device_model ON iot_device_bindings
    FOR EACH ROW
    EXECUTE FUNCTION increment_template_device_count();

-- 删除设备时，减少模板的device_count
CREATE OR REPLACE FUNCTION decrement_template_device_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE iot_sensor_templates
    SET device_count = device_count - 1
    WHERE model = OLD.device_model;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_decrement_device_count
    AFTER DELETE ON iot_device_bindings
    FOR EACH ROW
    EXECUTE FUNCTION decrement_template_device_count();

-- ===============================
-- 视图 - 设备完整信息（关联模板）
-- ===============================

CREATE OR REPLACE VIEW v_iot_device_bindings_with_template AS
SELECT
    b.id,
    b.device_id,
    b.device_name,
    b.device_alias,
    b.device_model,
    b.device_category,
    b.description,
    b.location,
    b.installation_date,
    b.status,
    b.is_online,
    b.online_status_updated_at,
    b.last_online_at,
    b.last_offline_at,
    b.last_data_at,
    b.tenant_id,
    b.org_id,
    b.created_by,
    b.updated_by,
    b.created_at,
    b.updated_at,
    -- 模板信息
    st.id AS template_id,
    st.name AS template_name,
    st.manufacturer AS template_manufacturer,
    st.enabled AS template_enabled
FROM iot_device_bindings b
LEFT JOIN iot_sensor_templates st ON b.device_model = st.model;

-- ===============================
-- 表注释
-- ===============================

COMMENT ON TABLE iot_sensor_templates IS 'IoT设备模板元数据表，详细配置存储在Etcd中，key格式: sensor-template-{id}';
COMMENT ON COLUMN iot_sensor_templates.id IS '模板UUID，用于构造Etcd配置key: sensor-template-{id}';
COMMENT ON COLUMN iot_sensor_templates.model IS '设备型号标识（业务标识，租户内唯一），如DHT22-v1';
COMMENT ON COLUMN iot_sensor_templates.device_count IS '使用该模板的设备数量（通过触发器自动维护）';

COMMENT ON TABLE iot_device_bindings IS 'IoT设备绑定表，整合设备基础信息和多租户、多组织绑定关系';
COMMENT ON COLUMN iot_device_bindings.device_id IS '设备唯一标识符，通常从MQTT消息的deviceIdPath提取';
COMMENT ON COLUMN iot_device_bindings.device_name IS '设备原始名称';
COMMENT ON COLUMN iot_device_bindings.device_alias IS '设备在组织内的别名，可自定义显示名称';
COMMENT ON COLUMN iot_device_bindings.device_model IS '设备型号，关联iot_sensor_templates表，模板中包含MQTT配置';
COMMENT ON COLUMN iot_device_bindings.is_online IS '在线状态，从Redis同步，通过后台任务定期更新';
COMMENT ON COLUMN iot_device_bindings.last_data_at IS '最后接收数据时间，每次收到MQTT消息时更新';
COMMENT ON COLUMN iot_device_bindings.tenant_id IS '所属租户ID，必填';
COMMENT ON COLUMN iot_device_bindings.org_id IS '所属组织ID，必填';

COMMENT ON TABLE iot_tag_definitions IS 'IoT设备标签定义表，租户隔离';
COMMENT ON COLUMN iot_tag_definitions.label IS '标签名称';
COMMENT ON COLUMN iot_tag_definitions.description IS '标签描述';
COMMENT ON COLUMN iot_tag_definitions.color IS '标签颜色（用于前端展示），如：#1890ff';
COMMENT ON COLUMN iot_tag_definitions.tenant_id IS '所属租户ID，标签在租户内隔离';

COMMENT ON TABLE iot_device_tag_relations IS 'IoT设备-标签关联表';
COMMENT ON COLUMN iot_device_tag_relations.device_id IS '设备ID，关联iot_device_bindings.device_id';
COMMENT ON COLUMN iot_device_tag_relations.tag_id IS '标签ID，关联iot_tag_definitions.id';

COMMENT ON TABLE iot_sensor_template_tag_relations IS '传感器模板-标签关联表';
COMMENT ON COLUMN iot_sensor_template_tag_relations.template_id IS '模板ID，关联iot_sensor_templates.id';
COMMENT ON COLUMN iot_sensor_template_tag_relations.tag_id IS '标签ID，关联iot_tag_definitions.id';

-- ===============================
-- 审计字段外键约束（在所有表创建后添加）
-- ===============================

-- iot_sensor_templates 表审计字段外键
ALTER TABLE iot_sensor_templates
    ADD CONSTRAINT fk_iot_sensor_templates_created_by
    FOREIGN KEY (created_by) REFERENCES system_users(id) ON DELETE SET NULL;

ALTER TABLE iot_sensor_templates
    ADD CONSTRAINT fk_iot_sensor_templates_updated_by
    FOREIGN KEY (updated_by) REFERENCES system_users(id) ON DELETE SET NULL;

-- iot_device_bindings 表审计字段外键
ALTER TABLE iot_device_bindings
    ADD CONSTRAINT fk_iot_device_bindings_created_by
    FOREIGN KEY (created_by) REFERENCES system_users(id) ON DELETE SET NULL;

ALTER TABLE iot_device_bindings
    ADD CONSTRAINT fk_iot_device_bindings_updated_by
    FOREIGN KEY (updated_by) REFERENCES system_users(id) ON DELETE SET NULL;

-- iot_tag_definitions 表审计字段外键
ALTER TABLE iot_tag_definitions
    ADD CONSTRAINT fk_iot_tag_definitions_created_by
    FOREIGN KEY (created_by) REFERENCES system_users(id) ON DELETE SET NULL;

ALTER TABLE iot_tag_definitions
    ADD CONSTRAINT fk_iot_tag_definitions_updated_by
    FOREIGN KEY (updated_by) REFERENCES system_users(id) ON DELETE SET NULL;
