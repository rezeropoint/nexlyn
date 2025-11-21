-- ===============================
-- Nexlyn Casbin权限管理系统表结构
-- 版本: v2.0 - 完整的RBAC权限控制系统
-- 支持功能: 基于角色的访问控制、多租户权限隔离
-- ===============================

-- Casbin权限规则表
CREATE TABLE IF NOT EXISTS casbin_rules (
    id BIGSERIAL PRIMARY KEY, -- Casbin GORM适配器需要整数类型ID
    ptype VARCHAR(10) NOT NULL, -- 策略类型 (p: 权限策略, g: 角色策略)
    v0 VARCHAR(100), -- 用户ID或角色名
    v1 VARCHAR(100), -- 租户ID或资源
    v2 VARCHAR(100), -- 资源或角色
    v3 VARCHAR(100), -- 动作
    v4 VARCHAR(100), -- 可选字段
    v5 VARCHAR(100), -- 可选字段
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ptype, v0, v1, v2, v3, v4, v5) -- 添加唯一约束防止重复规则
);

-- 为Casbin表创建索引
CREATE INDEX IF NOT EXISTS idx_casbin_ptype ON casbin_rules(ptype);
CREATE INDEX IF NOT EXISTS idx_casbin_v0 ON casbin_rules(v0);
CREATE INDEX IF NOT EXISTS idx_casbin_v1 ON casbin_rules(v1);
CREATE INDEX IF NOT EXISTS idx_casbin_v2 ON casbin_rules(v2);
CREATE INDEX IF NOT EXISTS idx_casbin_v0_v1 ON casbin_rules(v0, v1);
CREATE INDEX IF NOT EXISTS idx_casbin_updated_at ON casbin_rules(updated_at);

-- 为Casbin规则表创建更新时间触发器
CREATE TRIGGER update_casbin_rules_updated_at
    BEFORE UPDATE ON casbin_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ===============================
-- 角色定义表 - 用于区分角色和直接权限
-- ===============================

-- 角色定义表
CREATE TABLE IF NOT EXISTS system_roles (
    role_key VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    tenant_key VARCHAR(255) NOT NULL DEFAULT '*',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID, -- 创建者用户ID（UUID）
    updated_by UUID, -- 最后修改者用户ID（UUID）
    UNIQUE(role_key, tenant_key)
);

-- 为角色表创建索引
CREATE INDEX IF NOT EXISTS idx_system_roles_tenant_key ON system_roles(tenant_key);
CREATE INDEX IF NOT EXISTS idx_system_roles_created_at ON system_roles(created_at);
CREATE INDEX IF NOT EXISTS idx_system_roles_created_by ON system_roles(created_by);

-- 为角色表创建更新时间触发器
CREATE TRIGGER update_system_roles_updated_at
    BEFORE UPDATE ON system_roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ===============================
-- 数据库注释 - 权限管理表注释
-- ===============================

-- Casbin权限规则表注释
COMMENT ON TABLE casbin_rules IS 'Casbin权限管理规则表，存储RBAC权限配置';
COMMENT ON COLUMN casbin_rules.id IS '规则ID';
COMMENT ON COLUMN casbin_rules.ptype IS '策略类型：p-权限策略，g-角色策略';
COMMENT ON COLUMN casbin_rules.v0 IS '用户ID或角色名';
COMMENT ON COLUMN casbin_rules.v1 IS '租户ID或资源';
COMMENT ON COLUMN casbin_rules.v2 IS '资源或角色';
COMMENT ON COLUMN casbin_rules.v3 IS '操作动作';
COMMENT ON COLUMN casbin_rules.v4 IS '扩展字段4';
COMMENT ON COLUMN casbin_rules.v5 IS '扩展字段5';
COMMENT ON COLUMN casbin_rules.created_at IS '创建时间';
COMMENT ON COLUMN casbin_rules.updated_at IS '修改时间';

-- 角色定义表注释
COMMENT ON TABLE system_roles IS '角色定义表，用于区分角色和直接权限，支持多租户隔离';
COMMENT ON COLUMN system_roles.role_key IS '角色唯一标识';
COMMENT ON COLUMN system_roles.name IS '角色显示名称';
COMMENT ON COLUMN system_roles.description IS '角色描述';
COMMENT ON COLUMN system_roles.tenant_key IS '租户标识，*表示全局角色';
COMMENT ON COLUMN system_roles.created_at IS '创建时间';
COMMENT ON COLUMN system_roles.updated_at IS '修改时间';
COMMENT ON COLUMN system_roles.created_by IS '创建者用户ID';
COMMENT ON COLUMN system_roles.updated_by IS '最后修改者用户ID';

-- ===============================
-- 审计字段外键约束（在所有表创建后添加）
-- ===============================

-- system_roles 表审计字段外键
ALTER TABLE system_roles
    ADD CONSTRAINT fk_system_roles_created_by
    FOREIGN KEY (created_by) REFERENCES system_users(id) ON DELETE SET NULL;

ALTER TABLE system_roles
    ADD CONSTRAINT fk_system_roles_updated_by
    FOREIGN KEY (updated_by) REFERENCES system_users(id) ON DELETE SET NULL;
