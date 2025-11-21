-- ===============================
-- Nexlyn 系统初始化数据
-- 版本: v3.0 - 精简版（仅包含启动必需数据）
-- ===============================

-- 插入默认租户
INSERT INTO system_tenants (
    tenant_key, tenant_name, description, contact_email,
    status, max_users, created_by, updated_by
) VALUES (
    'default-tenant',
    '默认租户',
    '系统默认租户',
    'admin@nexlyn.com',
    'active',
    100,
    NULL,
    NULL
) ON CONFLICT (tenant_key) DO NOTHING;

-- 插入平台租户（用于平台级超级管理员归属）
INSERT INTO system_tenants (
    tenant_key, tenant_name, description, contact_email,
    status, max_users, created_by, updated_by
) VALUES (
    'platform',
    '平台租户',
    '平台级管理租户（系统保留）',
    'superadmin@nexlyn.com',
    'active',
    1000,
    NULL,
    NULL
) ON CONFLICT (tenant_key) DO NOTHING;

-- 插入默认管理员用户（密码：123456）
WITH default_tenant AS (
    SELECT id FROM system_tenants WHERE tenant_key = 'default-tenant'
)
INSERT INTO system_users (
    user_key, user_name, password_hash, name, email,
    tenant_id,
    status, created_by, updated_by
) VALUES (
    'admin-001',
    'admin',
    '$2a$10$UYR8aYmNiuJvn/9RZstgLO4dBmpvKUpR6gfjs93n1CHUoKMBlvAQm',
    '系统管理员',
    'admin@nexlyn.com',
    (SELECT id FROM default_tenant),
    'active', NULL, NULL
) ON CONFLICT (user_key) DO NOTHING;

-- 插入平台级超级管理员（密码：123456）
WITH platform_tenant AS (
    SELECT id FROM system_tenants WHERE tenant_key = 'platform'
)
INSERT INTO system_users (
    user_key, user_name, password_hash, name, email,
    tenant_id,
    status, created_by, updated_by
) VALUES (
    'super-admin-001',
    'superadmin',
    '$2a$10$UYR8aYmNiuJvn/9RZstgLO4dBmpvKUpR6gfjs93n1CHUoKMBlvAQm',
    '平台超级管理员',
    'superadmin@nexlyn.com',
    (SELECT id FROM platform_tenant),
    'active', NULL, NULL
) ON CONFLICT (user_key) DO NOTHING;

-- ===============================
-- 角色定义初始化
-- ===============================

-- 插入系统预定义角色
INSERT INTO system_roles (role_key, name, description, tenant_key, created_by, updated_by) VALUES
('super_admin', '平台超级管理员', '拥有所有权限，包括跨租户管理能力', '*',
 (SELECT id FROM system_users WHERE user_key = 'super-admin-001'),
 (SELECT id FROM system_users WHERE user_key = 'super-admin-001')),
('admin', '租户管理员', '租户内管理权限，包含系统管理功能', '*',
 (SELECT id FROM system_users WHERE user_key = 'super-admin-001'),
 (SELECT id FROM system_users WHERE user_key = 'super-admin-001')),
('user', '普通用户', '基础用户权限，只读访问', '*',
 (SELECT id FROM system_users WHERE user_key = 'super-admin-001'),
 (SELECT id FROM system_users WHERE user_key = 'super-admin-001'))
ON CONFLICT (role_key, tenant_key) DO NOTHING;

-- ===============================
-- Casbin 权限规则初始化
--
-- Casbin规则表字段说明：
-- ptype: 策略类型
--   - 'p': 权限策略 (policy)，定义主体对资源的操作权限
--   - 'g': 角色分组策略 (grouping)，定义用户与角色的关系
--
-- 权限策略 (ptype='p') 字段含义：
-- v0: 主体 (subject) - 用户ID或角色名
-- v1: 域 (domain) - 租户标识，'*'表示全局
-- v2: 资源 (resource) - 要访问的资源类型
-- v3: 操作 (action) - 对资源执行的操作
-- v4, v5: 扩展字段 (未使用)
--
-- 角色分组策略 (ptype='g') 字段含义：
-- v0: 用户ID (user)
-- v1: 角色名 (role)
-- v2: 域 (domain) - 角色生效的租户范围，'*'表示全局
-- v3, v4, v5: 扩展字段 (未使用)
--
-- 示例：
-- ('p', 'admin', '*', 'user', 'read') 表示：admin角色在所有租户下可以读取用户
-- ('g', 'user123', 'admin', 'tenant1') 表示：user123在tenant1中拥有admin角色
--
-- 注意：以下权限规则与 system_roles 表中定义的角色对应
-- ===============================

-- 角色权限策略 (p策略)
-- 超级管理员角色 - 跨租户全部权限（包括租户管理）
INSERT INTO casbin_rules (ptype, v0, v1, v2, v3) VALUES
('p', 'super_admin', '*', 'tenant', 'read'),
('p', 'super_admin', '*', 'tenant', 'write'),
('p', 'super_admin', '*', 'tenant', 'delete'),
('p', 'super_admin', '*', 'user', 'read'),
('p', 'super_admin', '*', 'user', 'write'),
('p', 'super_admin', '*', 'user', 'delete'),
('p', 'super_admin', '*', 'tag_user', 'read'),
('p', 'super_admin', '*', 'tag_user', 'write'),
('p', 'super_admin', '*', 'tag_user', 'delete'),
('p', 'super_admin', '*', 'tag_tenant', 'read'),
('p', 'super_admin', '*', 'tag_tenant', 'write'),
('p', 'super_admin', '*', 'tag_tenant', 'delete'),
('p', 'super_admin', '*', 'permission', 'read'),
('p', 'super_admin', '*', 'permission', 'write'),
('p', 'super_admin', '*', 'permission', 'delete'),
('p', 'super_admin', '*', 'role', 'read'),
('p', 'super_admin', '*', 'role', 'write'),
('p', 'super_admin', '*', 'role', 'delete'),
('p', 'super_admin', '*', 'system', 'read'),
('p', 'super_admin', '*', 'system', 'write'),
('p', 'super_admin', '*', 'system', 'delete'),
('p', 'super_admin', '*', 'gb28181_device', 'read'),
('p', 'super_admin', '*', 'gb28181_device', 'write'),
('p', 'super_admin', '*', 'gb28181_device', 'delete'),
('p', 'super_admin', '*', 'gb28181_stream', 'read'),
('p', 'super_admin', '*', 'gb28181_stream', 'write'),
('p', 'super_admin', '*', 'gb28181_stream', 'delete'),
('p', 'super_admin', '*', 'gb28181_tag', 'read'),
('p', 'super_admin', '*', 'gb28181_tag', 'write'),
('p', 'super_admin', '*', 'gb28181_tag', 'delete'),
('p', 'super_admin', '*', 'gb28181_stats', 'read'),
('p', 'super_admin', '*', 'organization', 'read'),
('p', 'super_admin', '*', 'organization', 'write'),
('p', 'super_admin', '*', 'organization', 'delete'),
('p', 'super_admin', '*', 'iot_template', 'read'),
('p', 'super_admin', '*', 'iot_template', 'write'),
('p', 'super_admin', '*', 'iot_template', 'delete'),
('p', 'super_admin', '*', 'iot_device', 'read'),
('p', 'super_admin', '*', 'iot_device', 'write'),
('p', 'super_admin', '*', 'iot_device', 'delete'),
('p', 'super_admin', '*', 'iot_tag', 'read'),
('p', 'super_admin', '*', 'iot_tag', 'write'),
('p', 'super_admin', '*', 'iot_tag', 'delete'),
('p', 'super_admin', '*', 'iot_metadata', 'read'),
('p', 'super_admin', '*', 'iot_metadata', 'write'),
('p', 'super_admin', '*', 'iot_metadata', 'delete'),
('p', 'super_admin', '*', 'iot_platform', 'read'),
('p', 'super_admin', '*', 'iot_platform', 'write'),
('p', 'super_admin', '*', 'iot_platform', 'delete'),
('p', 'super_admin', '*', 'skylark_platform', 'read'),
('p', 'super_admin', '*', 'skylark_platform', 'write'),
('p', 'super_admin', '*', 'skylark_platform', 'delete'),
('p', 'super_admin', '*', 'event_config', 'read'),
('p', 'super_admin', '*', 'event_config', 'write'),
('p', 'super_admin', '*', 'event_config', 'delete'),
('p', 'super_admin', '*', 'event_data', 'read'),
('p', 'super_admin', '*', 'event_data', 'write'),
('p', 'super_admin', '*', 'event_data', 'delete'),
('p', 'super_admin', '*', 'org_mapping', 'read'),
('p', 'super_admin', '*', 'org_mapping', 'write'),
('p', 'super_admin', '*', 'org_mapping', 'delete'),
('p', 'super_admin', '*', 'lynx_infoatomtype', 'read'),
('p', 'super_admin', '*', 'lynx_infoatomtype', 'write'),
('p', 'super_admin', '*', 'lynx_infoatomtype', 'delete'),
('p', 'super_admin', '*', 'lynx_graphconfig', 'read'),
('p', 'super_admin', '*', 'lynx_graphconfig', 'write'),
('p', 'super_admin', '*', 'lynx_graphconfig', 'delete'),
('p', 'super_admin', '*', 'lynx_blockspec', 'read'),
('p', 'super_admin', '*', 'lynx_tag', 'read'),
('p', 'super_admin', '*', 'lynx_tag', 'write'),
('p', 'super_admin', '*', 'lynx_tag', 'delete'),
('p', 'super_admin', '*', 'flow_journey', 'read'),
('p', 'super_admin', '*', 'flow_journey', 'write'),
('p', 'super_admin', '*', 'flow_journey', 'delete'),
('p', 'super_admin', '*', 'flow', 'read'),
('p', 'super_admin', '*', 'flow', 'write'),
('p', 'super_admin', '*', 'flow', 'delete'),
('p', 'super_admin', '*', 'form', 'write')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5) DO NOTHING;

-- 租户管理员角色 - 租户内管理权限（包含系统权限，使用全局域）
-- 注意：admin角色包含系统权限，因此会受到角色分配保护机制限制
INSERT INTO casbin_rules (ptype, v0, v1, v2, v3) VALUES
('p', 'admin', '*', 'user', 'read'),
('p', 'admin', '*', 'user', 'write'),
('p', 'admin', '*', 'user', 'delete'),
('p', 'admin', '*', 'tag_user', 'read'),
('p', 'admin', '*', 'tag_user', 'write'),
('p', 'admin', '*', 'tag_user', 'delete'),
('p', 'admin', '*', 'permission', 'read'),
('p', 'admin', '*', 'permission', 'write'),
('p', 'admin', '*', 'permission', 'delete'),
('p', 'admin', '*', 'role', 'read'),
('p', 'admin', '*', 'role', 'write'),
('p', 'admin', '*', 'role', 'delete'),
('p', 'admin', '*', 'system', 'read'),
('p', 'admin', '*', 'gb28181_device', 'read'),
('p', 'admin', '*', 'gb28181_device', 'write'),
('p', 'admin', '*', 'gb28181_device', 'delete'),
('p', 'admin', '*', 'gb28181_stream', 'read'),
('p', 'admin', '*', 'gb28181_stream', 'write'),
('p', 'admin', '*', 'gb28181_stream', 'delete'),
('p', 'admin', '*', 'gb28181_tag', 'read'),
('p', 'admin', '*', 'gb28181_tag', 'write'),
('p', 'admin', '*', 'gb28181_tag', 'delete'),
('p', 'admin', '*', 'gb28181_stats', 'read'),
('p', 'admin', '*', 'organization', 'read'),
('p', 'admin', '*', 'organization', 'write'),
('p', 'admin', '*', 'organization', 'delete'),
('p', 'admin', '*', 'iot_template', 'read'),
('p', 'admin', '*', 'iot_template', 'write'),
('p', 'admin', '*', 'iot_template', 'delete'),
('p', 'admin', '*', 'iot_device', 'read'),
('p', 'admin', '*', 'iot_device', 'write'),
('p', 'admin', '*', 'iot_device', 'delete'),
('p', 'admin', '*', 'iot_tag', 'read'),
('p', 'admin', '*', 'iot_tag', 'write'),
('p', 'admin', '*', 'iot_tag', 'delete'),
('p', 'admin', '*', 'iot_metadata', 'read'),
('p', 'admin', '*', 'iot_metadata', 'write'),
('p', 'admin', '*', 'iot_metadata', 'delete'),
('p', 'admin', '*', 'iot_platform', 'read'),
('p', 'admin', '*', 'iot_platform', 'write'),
('p', 'admin', '*', 'iot_platform', 'delete'),
('p', 'admin', '*', 'skylark_platform', 'read'),
('p', 'admin', '*', 'skylark_platform', 'write'),
('p', 'admin', '*', 'skylark_platform', 'delete'),
('p', 'admin', '*', 'event_config', 'read'),
('p', 'admin', '*', 'event_config', 'write'),
('p', 'admin', '*', 'event_config', 'delete'),
('p', 'admin', '*', 'event_data', 'read'),
('p', 'admin', '*', 'event_data', 'write'),
('p', 'admin', '*', 'event_data', 'delete'),
('p', 'admin', '*', 'org_mapping', 'read'),
('p', 'admin', '*', 'org_mapping', 'write'),
('p', 'admin', '*', 'org_mapping', 'delete'),
('p', 'admin', '*', 'lynx_infoatomtype', 'read'),
('p', 'admin', '*', 'lynx_infoatomtype', 'write'),
('p', 'admin', '*', 'lynx_infoatomtype', 'delete'),
('p', 'admin', '*', 'lynx_graphconfig', 'read'),
('p', 'admin', '*', 'lynx_graphconfig', 'write'),
('p', 'admin', '*', 'lynx_graphconfig', 'delete'),
('p', 'admin', '*', 'lynx_blockspec', 'read'),
('p', 'admin', '*', 'lynx_tag', 'read'),
('p', 'admin', '*', 'lynx_tag', 'write'),
('p', 'admin', '*', 'lynx_tag', 'delete'),
('p', 'admin', '*', 'flow_journey', 'read'),
('p', 'admin', '*', 'flow_journey', 'write'),
('p', 'admin', '*', 'flow_journey', 'delete'),
('p', 'admin', '*', 'flow', 'read'),
('p', 'admin', '*', 'flow', 'write'),
('p', 'admin', '*', 'flow', 'delete'),
('p', 'admin', '*', 'form', 'write')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5) DO NOTHING;

-- 普通用户角色 - 基础只读权限（租户级角色，无系统权限）
INSERT INTO casbin_rules (ptype, v0, v1, v2, v3) VALUES
('p', 'user', '*', 'user', 'read'),
('p', 'user', '*', 'tag_user', 'read'),
('p', 'user', '*', 'gb28181_device', 'read'),
('p', 'user', '*', 'gb28181_tag', 'read'),
('p', 'user', '*', 'organization', 'read'),
('p', 'user', '*', 'iot_template', 'read'),
('p', 'user', '*', 'iot_device', 'read'),
('p', 'user', '*', 'iot_tag', 'read'),
('p', 'user', '*', 'iot_metadata', 'read'),
('p', 'user', '*', 'iot_platform', 'read'),
('p', 'user', '*', 'lynx_infoatomtype', 'read'),
('p', 'user', '*', 'lynx_graphconfig', 'read'),
('p', 'user', '*', 'lynx_blockspec', 'read'),
('p', 'user', '*', 'lynx_tag', 'read'),
('p', 'user', '*', 'flow_journey', 'read'),
('p', 'user', '*', 'flow_journey', 'write'),
('p', 'user', '*', 'flow', 'read'),
('p', 'user', '*', 'form', 'write')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5) DO NOTHING;

-- 用户角色分配 (g策略)
-- 为默认管理员分配admin角色
-- 注意：admin角色包含系统权限，正常情况下不可分配，但租户初始化时例外
INSERT INTO casbin_rules (ptype, v0, v1, v2, v3) VALUES 
('g', 'admin-001', 'admin', 'default-tenant', '')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5) DO NOTHING;

-- 为超级管理员分配super_admin角色
INSERT INTO casbin_rules (ptype, v0, v1, v2, v3) VALUES
('g', 'super-admin-001', 'super_admin', '*', '')
ON CONFLICT (ptype, v0, v1, v2, v3, v4, v5) DO NOTHING;

-- ===============================
-- 默认组织架构数据初始化
-- ===============================

-- 插入默认租户的根组织
WITH default_tenant AS (
    SELECT id FROM system_tenants WHERE tenant_key = 'default-tenant'
)
INSERT INTO system_organizations (
    code, name, type, description,
    path, parent_id, level, sort_order,
    tenant_id, manager_id,
    status, created_by, updated_by
) VALUES (
    'ROOT',
    '默认组织',
    'company',
    '默认租户根组织',
    '/root/',
    NULL,
    1,
    0,
    (SELECT id FROM default_tenant),
    (SELECT id FROM system_users WHERE user_key = 'admin-001'),
    'active',
    (SELECT id FROM system_users WHERE user_key = 'admin-001'),
    (SELECT id FROM system_users WHERE user_key = 'admin-001')
) ON CONFLICT (code, tenant_id) DO NOTHING;

-- 插入平台租户的根组织
WITH platform_tenant AS (
    SELECT id FROM system_tenants WHERE tenant_key = 'platform'
)
INSERT INTO system_organizations (
    code, name, type, description,
    path, parent_id, level, sort_order,
    tenant_id, manager_id,
    status, created_by, updated_by
) VALUES (
    'PLATFORM_ROOT',
    '平台组织',
    'company',
    '平台租户根组织',
    '/root/',
    NULL,
    1,
    0,
    (SELECT id FROM platform_tenant),
    (SELECT id FROM system_users WHERE user_key = 'super-admin-001'),
    'active',
    (SELECT id FROM system_users WHERE user_key = 'super-admin-001'),
    (SELECT id FROM system_users WHERE user_key = 'super-admin-001')
) ON CONFLICT (code, tenant_id) DO NOTHING;

-- ===============================
-- 用户组织关联初始化
-- ===============================

-- 插入默认管理员组织关联
WITH default_tenant AS (
    SELECT id FROM system_tenants WHERE tenant_key = 'default-tenant'
), admin_user AS (
    SELECT id FROM system_users WHERE user_key = 'admin-001'
), root_org AS (
    SELECT id FROM system_organizations WHERE code = 'ROOT' AND tenant_id = (SELECT id FROM default_tenant)
)
INSERT INTO system_user_org_relations (
    user_id, org_id, relation_type, position_title, is_primary,
    status, created_by, updated_by
) VALUES
((SELECT id FROM admin_user), (SELECT id FROM root_org), 'manager', '管理员', TRUE, 'active',
 (SELECT id FROM admin_user), (SELECT id FROM admin_user))
ON CONFLICT (user_id, org_id) DO NOTHING;

-- 插入超级管理员组织关联
WITH platform_tenant AS (
    SELECT id FROM system_tenants WHERE tenant_key = 'platform'
), super_admin AS (
    SELECT id FROM system_users WHERE user_key = 'super-admin-001'
), platform_org AS (
    SELECT id FROM system_organizations WHERE code = 'PLATFORM_ROOT' AND tenant_id = (SELECT id FROM platform_tenant)
)
INSERT INTO system_user_org_relations (
    user_id, org_id, relation_type, position_title, is_primary,
    status, created_by, updated_by
) VALUES
((SELECT id FROM super_admin), (SELECT id FROM platform_org), 'manager', '超级管理员', TRUE, 'active',
 (SELECT id FROM super_admin), (SELECT id FROM super_admin))
ON CONFLICT (user_id, org_id) DO NOTHING;