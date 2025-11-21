package tag

// TableName LynxGraph标签定义表名
const TableName = "lynxgraph_tag_definitions"

// CreateTableSQL 创建Lynx标签定义表的PostgreSQL SQL语句
const CreateTableSQL = `
CREATE TABLE IF NOT EXISTS lynxgraph_tag_definitions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES system_tenants(id) ON DELETE CASCADE,
	name VARCHAR(100) NOT NULL,
	description TEXT,
	scope VARCHAR(20) NOT NULL,
	created_by VARCHAR(100),
	updated_by VARCHAR(100),
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT uk_lynxgraph_tag_name_tenant_scope UNIQUE(name, tenant_id, scope),
	CONSTRAINT ck_lynxgraph_tag_scope CHECK(scope IN ('info_atom', 'logic_graph'))
);

-- 索引优化
CREATE INDEX IF NOT EXISTS idx_lynxgraph_tag_defs_tenant ON lynxgraph_tag_definitions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_lynxgraph_tag_defs_tenant_scope ON lynxgraph_tag_definitions(tenant_id, scope);
CREATE INDEX IF NOT EXISTS idx_lynxgraph_tag_defs_tenant_name ON lynxgraph_tag_definitions(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_lynxgraph_tag_defs_created_at ON lynxgraph_tag_definitions(created_at DESC);

-- 自动更新updated_at字段的触发器
CREATE OR REPLACE FUNCTION update_lynxgraph_tag_definitions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_lynxgraph_tag_definitions_updated_at ON lynxgraph_tag_definitions;
CREATE TRIGGER trigger_lynxgraph_tag_definitions_updated_at
    BEFORE UPDATE ON lynxgraph_tag_definitions
    FOR EACH ROW
    EXECUTE FUNCTION update_lynxgraph_tag_definitions_updated_at();
`

// CheckTableExistsSQL 检查lynxgraph_tag_definitions表是否存在
const CheckTableExistsSQL = `SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'lynxgraph_tag_definitions'`
