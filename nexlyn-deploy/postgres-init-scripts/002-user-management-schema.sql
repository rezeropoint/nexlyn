-- ===============================
-- Nexlyn 多租户用户管理系统数据库表结构
-- 版本: v2.0 - 包含完整的租户管理、权限管理、标签系统
-- 支持功能: 多租户隔离、RBAC权限控制、标签分类、软删除
-- ===============================

-- 租户表
CREATE TABLE IF NOT EXISTS system_tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_key VARCHAR(100) UNIQUE NOT NULL, -- 租户唯一标识符，增加长度
    tenant_name VARCHAR(200) NOT NULL, -- 租户名称
    description TEXT, -- 租户描述
    contact_email VARCHAR(255), -- 联系邮箱
    contact_phone VARCHAR(20), -- 联系电话
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'deleted')), -- 租户状态：active-活跃，inactive-禁用，deleted-删除
    settings JSONB DEFAULT '{}', -- 租户设置（JSON格式）
    max_users INTEGER DEFAULT 100 CHECK (max_users > 0), -- 用户数量限制
    expires_at TIMESTAMP WITH TIME ZONE, -- 过期时间
    created_by UUID, -- 创建者用户ID（UUID）
    updated_by UUID, -- 最后修改者用户ID（UUID）
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE -- 软删除时间
);

-- 用户主表
CREATE TABLE IF NOT EXISTS system_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- 用户UUID, 主键
    user_key VARCHAR(100) UNIQUE NOT NULL, -- 用户唯一标识符，例如 admin-001
    user_name VARCHAR(100) NOT NULL, -- 用户名，用于登录
    password_hash VARCHAR(255) NOT NULL, -- 密码哈希
    name VARCHAR(100) NOT NULL, -- 用户姓名
    email VARCHAR(255) NOT NULL, -- 电子邮件
    phone VARCHAR(20), -- 电话号码
    avatar VARCHAR(500), -- 头像URL
    signature TEXT, -- 个人签名
    title VARCHAR(100), -- 职位

    -- 多租户信息
    tenant_id UUID REFERENCES system_tenants(id) ON DELETE CASCADE, -- 所属租户


    -- 地理信息
    country VARCHAR(100), -- 国家
    province VARCHAR(100), -- 省份
    city VARCHAR(100), -- 城市
    address TEXT, -- 地址

    -- 状态和时间戳
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'deleted')), -- 用户状态：active-活跃，inactive-禁用，deleted-删除
    created_by UUID, -- 创建者用户ID（UUID）
    updated_by UUID, -- 最后修改者用户ID（UUID）
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE, -- 软删除时间

    -- 唯一约束：用户名全局唯一，邮箱在租户内唯一
    UNIQUE(user_name),
    UNIQUE(email, tenant_id)
);



-- 用户地理信息表
CREATE TABLE IF NOT EXISTS user_geographic (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES system_users(id) ON DELETE CASCADE,
    province_key VARCHAR(50),
    province_label VARCHAR(100),
    city_key VARCHAR(50),
    city_label VARCHAR(100),
    district_key VARCHAR(50), -- 区县
    district_label VARCHAR(100), -- 区县名称
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);



-- 标签定义表（字典），按作用域区分（tenant | user）
CREATE TABLE IF NOT EXISTS system_tag_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope VARCHAR(20) NOT NULL, -- 'tenant' | 'user'
    label VARCHAR(100) NOT NULL,
    description TEXT,
    tenant_id UUID REFERENCES system_tenants(id) ON DELETE CASCADE, -- 租户隔离：user类型tag需要指定所属租户，tenant类型tag为NULL（全局）
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- 租户标签在全局范围内唯一，用户标签在租户内唯一
    UNIQUE(scope, label, tenant_id)
);


CREATE INDEX IF NOT EXISTS idx_system_tag_defs_scope_label ON system_tag_definitions(scope, label);

-- 为标签表添加更多索引
CREATE INDEX IF NOT EXISTS idx_system_tag_defs_scope ON system_tag_definitions(scope);
CREATE INDEX IF NOT EXISTS idx_system_tag_defs_tenant_id ON system_tag_definitions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_system_tag_defs_scope_tenant ON system_tag_definitions(scope, tenant_id);
CREATE INDEX IF NOT EXISTS idx_system_tag_defs_created_at ON system_tag_definitions(created_at);

-- 用户-标签关联表（使用标签ID进行关联）
CREATE TABLE IF NOT EXISTS user_tag_relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES system_users(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES system_tag_definitions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_user_tag_rel_user_id ON user_tag_relations(user_id);
CREATE INDEX IF NOT EXISTS idx_user_tag_rel_tag_id ON user_tag_relations(tag_id);
CREATE INDEX IF NOT EXISTS idx_user_tag_rel_created_at ON user_tag_relations(created_at);
CREATE INDEX IF NOT EXISTS idx_user_tag_rel_updated_at ON user_tag_relations(updated_at);

-- 租户-标签关联表（使用标签ID进行关联）
CREATE TABLE IF NOT EXISTS system_tenant_tag_relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES system_tenants(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES system_tag_definitions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_system_tenant_tag_rel_tenant_id ON system_tenant_tag_relations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_system_tenant_tag_rel_tag_id ON system_tenant_tag_relations(tag_id);
CREATE INDEX IF NOT EXISTS idx_system_tenant_tag_rel_created_at ON system_tenant_tag_relations(created_at);
CREATE INDEX IF NOT EXISTS idx_system_tenant_tag_rel_updated_at ON system_tenant_tag_relations(updated_at);

-- 创建索引提高查询性能
-- 租户表索引
CREATE INDEX IF NOT EXISTS idx_system_tenants_tenant_key ON system_tenants(tenant_key);
CREATE INDEX IF NOT EXISTS idx_system_tenants_status ON system_tenants(status);
CREATE INDEX IF NOT EXISTS idx_system_tenants_created_by ON system_tenants(created_by);
CREATE INDEX IF NOT EXISTS idx_system_tenants_expires_at ON system_tenants(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_system_tenants_deleted_at ON system_tenants(deleted_at) WHERE deleted_at IS NOT NULL;

-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_system_users_user_key ON system_users(user_key);
CREATE INDEX IF NOT EXISTS idx_system_users_username ON system_users(user_name);
CREATE INDEX IF NOT EXISTS idx_system_users_email_tenant ON system_users(email, tenant_id);
CREATE INDEX IF NOT EXISTS idx_system_users_tenant_id ON system_users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_system_users_created_at ON system_users(created_at);
CREATE INDEX IF NOT EXISTS idx_system_users_status ON system_users(status);
CREATE INDEX IF NOT EXISTS idx_system_users_last_login ON system_users(last_login_at) WHERE last_login_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_system_users_deleted_at ON system_users(deleted_at) WHERE deleted_at IS NOT NULL;




-- 用户地理信息表索引
CREATE INDEX IF NOT EXISTS idx_user_geographic_user_id ON user_geographic(user_id);
CREATE INDEX IF NOT EXISTS idx_user_geographic_province ON user_geographic(province_key);
CREATE INDEX IF NOT EXISTS idx_user_geographic_city ON user_geographic(city_key);
CREATE INDEX IF NOT EXISTS idx_user_geographic_district ON user_geographic(district_key);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为租户表创建更新时间触发器
CREATE TRIGGER update_system_tenants_updated_at
    BEFORE UPDATE ON system_tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 为用户表创建更新时间触发器
CREATE TRIGGER update_system_users_updated_at
    BEFORE UPDATE ON system_users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 为用户地理信息表创建更新时间触发器
CREATE TRIGGER update_user_geographic_updated_at
    BEFORE UPDATE ON user_geographic
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 为标签定义表创建更新时间触发器
CREATE TRIGGER update_system_tag_definitions_updated_at
    BEFORE UPDATE ON system_tag_definitions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 为用户标签关联表创建更新时间触发器
CREATE TRIGGER update_user_tag_relations_updated_at
    BEFORE UPDATE ON user_tag_relations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 为租户标签关联表创建更新时间触发器
CREATE TRIGGER update_system_tenant_tag_relations_updated_at
    BEFORE UPDATE ON system_tenant_tag_relations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ===============================
-- 数据库注释 - 用户管理相关表注释
-- ===============================

-- 租户表注释
COMMENT ON TABLE system_tenants IS '多租户系统的租户表，用于管理不同的租户组织';
COMMENT ON COLUMN system_tenants.id IS '租户UUID主键';
COMMENT ON COLUMN system_tenants.tenant_key IS '租户唯一标识符，用于API调用';
COMMENT ON COLUMN system_tenants.tenant_name IS '租户名称';
COMMENT ON COLUMN system_tenants.description IS '租户描述信息';
COMMENT ON COLUMN system_tenants.contact_email IS '租户联系邮箱';
COMMENT ON COLUMN system_tenants.contact_phone IS '租户联系电话';
COMMENT ON COLUMN system_tenants.status IS '租户状态：active-活跃，inactive-禁用，deleted-删除';
COMMENT ON COLUMN system_tenants.settings IS '租户配置设置，JSON格式存储';
COMMENT ON COLUMN system_tenants.max_users IS '租户用户数量限制';
COMMENT ON COLUMN system_tenants.expires_at IS '租户过期时间';
COMMENT ON COLUMN system_tenants.created_by IS '创建者用户ID';
COMMENT ON COLUMN system_tenants.updated_by IS '最后修改者用户ID';
COMMENT ON COLUMN system_tenants.created_at IS '创建时间';
COMMENT ON COLUMN system_tenants.updated_at IS '最后修改时间';
COMMENT ON COLUMN system_tenants.deleted_at IS '软删除时间';

-- 用户表注释
COMMENT ON TABLE system_users IS '用户信息表，支持多租户隔离';
COMMENT ON COLUMN system_users.id IS '用户UUID主键';
COMMENT ON COLUMN system_users.user_key IS '用户唯一标识符';
COMMENT ON COLUMN system_users.user_name IS '用户名，用于登录';
COMMENT ON COLUMN system_users.password_hash IS '密码哈希值';
COMMENT ON COLUMN system_users.name IS '用户真实姓名';
COMMENT ON COLUMN system_users.email IS '用户邮箱';
COMMENT ON COLUMN system_users.phone IS '用户电话号码';
COMMENT ON COLUMN system_users.avatar IS '用户头像URL';
COMMENT ON COLUMN system_users.signature IS '用户个人签名';
COMMENT ON COLUMN system_users.title IS '用户职位';
COMMENT ON COLUMN system_users.tenant_id IS '所属租户ID';
COMMENT ON COLUMN system_users.country IS '用户所在国家';
COMMENT ON COLUMN system_users.province IS '用户所在省份';
COMMENT ON COLUMN system_users.city IS '用户所在城市';
COMMENT ON COLUMN system_users.address IS '用户详细地址';
COMMENT ON COLUMN system_users.status IS '用户状态：active-活跃，inactive-禁用，deleted-删除';
COMMENT ON COLUMN system_users.created_at IS '创建时间';
COMMENT ON COLUMN system_users.updated_at IS '最后修改时间';
COMMENT ON COLUMN system_users.last_login_at IS '最后登录时间';
COMMENT ON COLUMN system_users.deleted_at IS '软删除时间';

-- 用户地理信息表注释
COMMENT ON TABLE user_geographic IS '用户地理位置信息表';
COMMENT ON COLUMN user_geographic.id IS '主键ID';
COMMENT ON COLUMN user_geographic.user_id IS '关联用户ID';
COMMENT ON COLUMN user_geographic.province_key IS '省份代码';
COMMENT ON COLUMN user_geographic.province_label IS '省份名称';
COMMENT ON COLUMN user_geographic.city_key IS '城市代码';
COMMENT ON COLUMN user_geographic.city_label IS '城市名称';
COMMENT ON COLUMN user_geographic.district_key IS '区县代码';
COMMENT ON COLUMN user_geographic.district_label IS '区县名称';
COMMENT ON COLUMN user_geographic.created_at IS '创建时间';
COMMENT ON COLUMN user_geographic.updated_at IS '修改时间';

-- 标签定义表注释
COMMENT ON TABLE system_tag_definitions IS '标签定义表，支持租户级和用户级标签';
COMMENT ON COLUMN system_tag_definitions.id IS '标签定义ID';
COMMENT ON COLUMN system_tag_definitions.scope IS '标签作用域：tenant-租户标签，user-用户标签';
COMMENT ON COLUMN system_tag_definitions.label IS '标签名称';
COMMENT ON COLUMN system_tag_definitions.description IS '标签描述';
COMMENT ON COLUMN system_tag_definitions.tenant_id IS '所属租户ID，租户标签为NULL（全局）';
COMMENT ON COLUMN system_tag_definitions.created_at IS '创建时间';
COMMENT ON COLUMN system_tag_definitions.updated_at IS '修改时间';

-- 用户标签关联表注释
COMMENT ON TABLE user_tag_relations IS '用户和标签的关联表';
COMMENT ON COLUMN user_tag_relations.id IS '关联记录ID';
COMMENT ON COLUMN user_tag_relations.user_id IS '用户ID';
COMMENT ON COLUMN user_tag_relations.tag_id IS '标签定义ID';
COMMENT ON COLUMN user_tag_relations.created_at IS '创建时间';
COMMENT ON COLUMN user_tag_relations.updated_at IS '修改时间';

-- 租户标签关联表注释
COMMENT ON TABLE system_tenant_tag_relations IS '租户和标签的关联表';
COMMENT ON COLUMN system_tenant_tag_relations.id IS '关联记录ID';
COMMENT ON COLUMN system_tenant_tag_relations.tenant_id IS '租户ID';
COMMENT ON COLUMN system_tenant_tag_relations.tag_id IS '标签定义ID';
COMMENT ON COLUMN system_tenant_tag_relations.created_at IS '创建时间';
COMMENT ON COLUMN system_tenant_tag_relations.updated_at IS '修改时间';

-- ===============================
-- 审计字段外键约束（在所有表创建后添加）
-- ===============================

-- system_tenants 表审计字段外键
ALTER TABLE system_tenants
    ADD CONSTRAINT fk_system_tenants_created_by
    FOREIGN KEY (created_by) REFERENCES system_users(id) ON DELETE SET NULL;

ALTER TABLE system_tenants
    ADD CONSTRAINT fk_system_tenants_updated_by
    FOREIGN KEY (updated_by) REFERENCES system_users(id) ON DELETE SET NULL;

-- system_users 表审计字段外键（自引用）
ALTER TABLE system_users
    ADD CONSTRAINT fk_system_users_created_by
    FOREIGN KEY (created_by) REFERENCES system_users(id) ON DELETE SET NULL;

ALTER TABLE system_users
    ADD CONSTRAINT fk_system_users_updated_by
    FOREIGN KEY (updated_by) REFERENCES system_users(id) ON DELETE SET NULL;