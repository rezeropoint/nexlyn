-- ===============================
-- Nexlyn 路径编码模式组织架构系统数据库表结构
-- 版本: v1.0 - 基于路径编码的树形组织架构管理
-- 支持功能: 多租户组织隔离、路径编码树形结构、快速层级查询、矩阵式用户管理
-- 设计模式: Path Enumeration Pattern (路径枚举模式)
-- ===============================

-- 组织架构表（使用路径编码模式）
CREATE TABLE IF NOT EXISTS system_organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL, -- 组织代码，租户内唯一
    name VARCHAR(200) NOT NULL, -- 组织名称
    type VARCHAR(50) DEFAULT 'department' CHECK (type IN ('company', 'department', 'team', 'group')), -- 组织类型
    description TEXT, -- 组织描述

    -- 路径编码核心字段
    path VARCHAR(1000) NOT NULL, -- 路径编码：/root/parent_id/current_id/ 格式
    parent_id UUID REFERENCES system_organizations(id) ON DELETE CASCADE, -- 父组织ID（冗余字段，便于快速查询）
    level INTEGER NOT NULL DEFAULT 1 CHECK (level > 0 AND level <= 10), -- 层级深度，限制最大10层
    sort_order INTEGER DEFAULT 0, -- 同级排序序号

    -- 多租户支持
    tenant_id UUID REFERENCES system_tenants(id) ON DELETE CASCADE,

    -- 组织负责人信息
    manager_id UUID REFERENCES system_users(id) ON DELETE SET NULL, -- 组织负责人
    contact_email VARCHAR(255), -- 组织联系邮箱
    contact_phone VARCHAR(20), -- 组织联系电话


    -- 状态和时间戳
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'deleted')), -- 组织状态
    created_by UUID, -- 创建者用户ID（UUID）
    updated_by UUID, -- 最后修改者用户ID（UUID）
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE, -- 软删除时间

    -- 约束条件
    UNIQUE(code, tenant_id), -- 组织代码在租户内唯一
    UNIQUE(path, tenant_id) -- 路径在租户内唯一
);

-- 用户-组织关联表（支持矩阵式管理）
CREATE TABLE IF NOT EXISTS system_user_org_relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES system_users(id) ON DELETE CASCADE,
    org_id UUID REFERENCES system_organizations(id) ON DELETE CASCADE,
    
    -- 关联属性
    relation_type VARCHAR(20) DEFAULT 'member' CHECK (relation_type IN ('manager', 'member', 'deputy', 'consultant')), -- 关系类型
    position_title VARCHAR(100), -- 在该组织的职位名称
    is_primary BOOLEAN DEFAULT FALSE, -- 是否为主属组织
    start_date DATE, -- 开始日期
    end_date DATE, -- 结束日期（为空表示当前有效）
    
    -- 权限相关
    permissions JSONB DEFAULT '{}', -- 在该组织的特殊权限（JSON格式）
    
    -- 状态和时间戳
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'deleted')),
    created_by UUID, -- 创建者用户ID（UUID）
    updated_by UUID, -- 最后修改者用户ID（UUID）
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- 约束条件
    UNIQUE(user_id, org_id), -- 用户在同一组织只能有一条有效记录
    CHECK(end_date IS NULL OR end_date >= start_date) -- 结束日期不能早于开始日期
);

-- ===============================
-- 路径编码查询优化索引
-- ===============================

-- 组织表核心索引
CREATE INDEX IF NOT EXISTS idx_system_organizations_path ON system_organizations(path);
CREATE INDEX IF NOT EXISTS idx_system_organizations_path_pattern ON system_organizations(path text_pattern_ops); -- 支持LIKE查询优化
CREATE INDEX IF NOT EXISTS idx_system_organizations_tenant_path ON system_organizations(tenant_id, path);
CREATE INDEX IF NOT EXISTS idx_system_organizations_parent_id ON system_organizations(parent_id);
CREATE INDEX IF NOT EXISTS idx_system_organizations_tenant_id ON system_organizations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_system_organizations_level ON system_organizations(level);
CREATE INDEX IF NOT EXISTS idx_system_organizations_code_tenant ON system_organizations(code, tenant_id);
CREATE INDEX IF NOT EXISTS idx_system_organizations_manager_id ON system_organizations(manager_id) WHERE manager_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_system_organizations_status ON system_organizations(status);
CREATE INDEX IF NOT EXISTS idx_system_organizations_created_at ON system_organizations(created_at);
CREATE INDEX IF NOT EXISTS idx_system_organizations_deleted_at ON system_organizations(deleted_at) WHERE deleted_at IS NOT NULL;

-- 用户组织关联表索引
CREATE INDEX IF NOT EXISTS idx_system_user_org_rel_user_id ON system_user_org_relations(user_id);
CREATE INDEX IF NOT EXISTS idx_system_user_org_rel_org_id ON system_user_org_relations(org_id);
CREATE INDEX IF NOT EXISTS idx_system_user_org_rel_user_org ON system_user_org_relations(user_id, org_id);
CREATE INDEX IF NOT EXISTS idx_system_user_org_rel_is_primary ON system_user_org_relations(is_primary) WHERE is_primary = TRUE;
CREATE INDEX IF NOT EXISTS idx_system_user_org_rel_relation_type ON system_user_org_relations(relation_type);
CREATE INDEX IF NOT EXISTS idx_system_user_org_rel_status ON system_user_org_relations(status);
CREATE INDEX IF NOT EXISTS idx_system_user_org_rel_active_period ON system_user_org_relations(start_date, end_date) WHERE end_date IS NULL;

-- ===============================
-- 数据完整性约束和触发器
-- ===============================

-- 组织表更新时间触发器
CREATE TRIGGER update_system_organizations_updated_at
    BEFORE UPDATE ON system_organizations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 用户组织关联表更新时间触发器
CREATE TRIGGER update_system_user_org_relations_updated_at
    BEFORE UPDATE ON system_user_org_relations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 路径编码一致性触发器函数
CREATE OR REPLACE FUNCTION maintain_organization_path()
RETURNS TRIGGER AS $$
DECLARE
    parent_path VARCHAR(1000);
    new_path VARCHAR(1000);
BEGIN
    -- 插入或更新时维护路径一致性
    IF NEW.parent_id IS NULL THEN
        -- 根节点，路径为 /root/
        NEW.path := '/root/';
        NEW.level := 1;
    ELSE
        -- 子节点，获取父节点路径
        SELECT path, level INTO parent_path, NEW.level
        FROM system_organizations
        WHERE id = NEW.parent_id AND tenant_id = NEW.tenant_id;
        
        IF parent_path IS NULL THEN
            RAISE EXCEPTION '无效的父组织ID: %', NEW.parent_id;
        END IF;
        
        -- 构建新路径：父路径 + 当前ID + /
        NEW.path := parent_path || NEW.id || '/';
        NEW.level := NEW.level + 1;
        
        -- 检查层级限制
        IF NEW.level > 10 THEN
            RAISE EXCEPTION '组织层级不能超过10层';
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建路径维护触发器
CREATE TRIGGER trigger_maintain_organization_path
    BEFORE INSERT OR UPDATE ON system_organizations
    FOR EACH ROW
    EXECUTE FUNCTION maintain_organization_path();

-- ===============================
-- 路径编码辅助函数
-- ===============================

-- ===============================
-- 数据库注释 - 组织架构相关表注释
-- 组织架构表注释
COMMENT ON TABLE system_organizations IS '路径编码模式的组织架构表，支持多租户树形结构管理';
COMMENT ON COLUMN system_organizations.id IS '组织UUID主键';
COMMENT ON COLUMN system_organizations.code IS '组织代码，租户内唯一标识';
COMMENT ON COLUMN system_organizations.name IS '组织名称';
COMMENT ON COLUMN system_organizations.type IS '组织类型：company-公司，department-部门，team-团队，group-小组';
COMMENT ON COLUMN system_organizations.description IS '组织描述信息';
COMMENT ON COLUMN system_organizations.path IS '路径编码字符串，格式：/root/parent_id/current_id/';
COMMENT ON COLUMN system_organizations.parent_id IS '父组织ID，冗余字段用于快速查询';
COMMENT ON COLUMN system_organizations.level IS '组织层级深度，从1开始，最大10层';
COMMENT ON COLUMN system_organizations.sort_order IS '同级组织排序序号';
COMMENT ON COLUMN system_organizations.tenant_id IS '所属租户ID，支持多租户隔离';
COMMENT ON COLUMN system_organizations.manager_id IS '组织负责人用户ID';
COMMENT ON COLUMN system_organizations.contact_email IS '组织联系邮箱';
COMMENT ON COLUMN system_organizations.contact_phone IS '组织联系电话';
COMMENT ON COLUMN system_organizations.status IS '组织状态：active-活跃，inactive-停用，deleted-删除';
COMMENT ON COLUMN system_organizations.created_by IS '创建者用户ID';
COMMENT ON COLUMN system_organizations.updated_by IS '最后修改者用户ID';
COMMENT ON COLUMN system_organizations.created_at IS '创建时间';
COMMENT ON COLUMN system_organizations.updated_at IS '最后修改时间';
COMMENT ON COLUMN system_organizations.deleted_at IS '软删除时间';

-- 用户组织关联表注释
COMMENT ON TABLE system_user_org_relations IS '用户和组织的关联表，支持矩阵式组织管理';
COMMENT ON COLUMN system_user_org_relations.id IS '关联记录UUID主键';
COMMENT ON COLUMN system_user_org_relations.user_id IS '用户ID';
COMMENT ON COLUMN system_user_org_relations.org_id IS '组织ID';
COMMENT ON COLUMN system_user_org_relations.relation_type IS '关系类型：manager-管理者，member-成员，deputy-副职，consultant-顾问';
COMMENT ON COLUMN system_user_org_relations.position_title IS '用户在该组织的职位名称';
COMMENT ON COLUMN system_user_org_relations.is_primary IS '是否为用户的主属组织';
COMMENT ON COLUMN system_user_org_relations.start_date IS '关联开始日期';
COMMENT ON COLUMN system_user_org_relations.end_date IS '关联结束日期，为NULL表示当前有效';
COMMENT ON COLUMN system_user_org_relations.permissions IS '用户在该组织的特殊权限配置（JSON格式）';
COMMENT ON COLUMN system_user_org_relations.status IS '关联状态：active-活跃，inactive-停用，deleted-删除';
COMMENT ON COLUMN system_user_org_relations.created_by IS '创建者用户ID';
COMMENT ON COLUMN system_user_org_relations.updated_by IS '最后修改者用户ID';
COMMENT ON COLUMN system_user_org_relations.created_at IS '创建时间';
COMMENT ON COLUMN system_user_org_relations.updated_at IS '最后修改时间';
COMMENT ON COLUMN system_user_org_relations.deleted_at IS '软删除时间';

-- 辅助函数注释
COMMENT ON FUNCTION maintain_organization_path() IS '维护组织路径编码一致性的触发器函数';

-- ===============================
-- 路径编码模式使用示例
-- ===============================

/*
路径编码模式使用示例：

1. 创建组织结构：
   根组织：path = '/root/'
   一级部门：path = '/root/dept1_uuid/'
   二级部门：path = '/root/dept1_uuid/subdept1_uuid/'

2. 快速查询所有子组织：
   SELECT * FROM system_organizations WHERE path LIKE '/root/dept1_uuid/%' AND path != '/root/dept1_uuid/';

3. 获取组织层级：
   SELECT level FROM system_organizations WHERE id = 'org_uuid';

4. 查询特定组织下的所有子孙组织：
   SELECT * FROM system_organizations WHERE path LIKE '/root/parent_uuid/%' AND path != '/root/parent_uuid/';

路径编码模式优势：
- 快速查询子树：通过路径前缀匹配
- 快速计算层级：通过路径解析
- 支持高效的树形查询和操作
- 避免递归查询的性能问题
- 支持多租户隔离

性能特点：
- 查询子孙节点：O(n) 其中n为结果集大小，无需递归
- 查询祖先节点：O(d) 其中d为层级深度
- 插入节点：O(1) 只需计算路径
- 移动子树：O(m) 其中m为子树节点数量
- 空间复杂度：路径字符串长度约为 32*层级深度 字节
*/

-- ===============================
-- 审计字段外键约束（在所有表创建后添加）
-- ===============================

-- system_organizations 表审计字段外键
ALTER TABLE system_organizations
    ADD CONSTRAINT fk_system_organizations_created_by
    FOREIGN KEY (created_by) REFERENCES system_users(id) ON DELETE SET NULL;

ALTER TABLE system_organizations
    ADD CONSTRAINT fk_system_organizations_updated_by
    FOREIGN KEY (updated_by) REFERENCES system_users(id) ON DELETE SET NULL;

-- system_user_org_relations 表审计字段外键
ALTER TABLE system_user_org_relations
    ADD CONSTRAINT fk_system_user_org_relations_created_by
    FOREIGN KEY (created_by) REFERENCES system_users(id) ON DELETE SET NULL;

ALTER TABLE system_user_org_relations
    ADD CONSTRAINT fk_system_user_org_relations_updated_by
    FOREIGN KEY (updated_by) REFERENCES system_users(id) ON DELETE SET NULL;