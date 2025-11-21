package graph

// TableName LynxGraph逻辑图配置表名
const TableName = "lynxgraph_graph_configs"

// CreateTableSQL 创建逻辑图表的PostgreSQL SQL语句
const CreateTableSQL = `
CREATE TABLE IF NOT EXISTS lynxgraph_graph_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES system_tenants(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES system_organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(32) NOT NULL,
    description TEXT,
    tag_ids UUID[] DEFAULT '{}'::UUID[],
    enable BOOLEAN NOT NULL DEFAULT true,
    icon VARCHAR(50),
    icon_color VARCHAR(50),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_lynxgraph_graph_configs_tenant_name_version UNIQUE(tenant_id, name, version)
);

-- 索引优化
CREATE INDEX IF NOT EXISTS idx_lynxgraph_graph_configs_tenant ON lynxgraph_graph_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_lynxgraph_graph_configs_org ON lynxgraph_graph_configs(org_id);
CREATE INDEX IF NOT EXISTS idx_lynxgraph_graph_configs_tenant_org ON lynxgraph_graph_configs(tenant_id, org_id);
CREATE INDEX IF NOT EXISTS idx_lynxgraph_graph_configs_name ON lynxgraph_graph_configs(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_lynxgraph_graph_configs_enabled ON lynxgraph_graph_configs(tenant_id, enable);
CREATE INDEX IF NOT EXISTS idx_lynxgraph_graph_configs_created_at ON lynxgraph_graph_configs(created_at DESC);

-- 复合索引（查询已启用的图）
CREATE INDEX IF NOT EXISTS idx_lynxgraph_graph_configs_tenant_enabled ON lynxgraph_graph_configs(tenant_id, enable);

-- 自动更新updated_at字段的触发器
CREATE OR REPLACE FUNCTION update_lynxgraph_graph_configs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_lynxgraph_graph_configs_updated_at ON lynxgraph_graph_configs;
CREATE TRIGGER trigger_lynxgraph_graph_configs_updated_at
    BEFORE UPDATE ON lynxgraph_graph_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_lynxgraph_graph_configs_updated_at();
`

// CheckTableExistsSQL 检查lynxgraph_graph_configs表是否存在
const CheckTableExistsSQL = `SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'lynxgraph_graph_configs'`

// CreateTagAssociationTableSQL 创建逻辑图标签关联表的PostgreSQL SQL语句
const CreateTagAssociationTableSQL = `
CREATE TABLE IF NOT EXISTS graph_config_tags (
    graph_config_id UUID NOT NULL REFERENCES lynxgraph_graph_configs(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES lynxgraph_tag_definitions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (graph_config_id, tag_id)
);

-- 索引优化
CREATE INDEX IF NOT EXISTS idx_graph_config_tags_graph ON graph_config_tags(graph_config_id);
CREATE INDEX IF NOT EXISTS idx_graph_config_tags_tag ON graph_config_tags(tag_id);
`
