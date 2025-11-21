package infoatom

// CreateTableSQL 创建信息原子类型表的PostgreSQL SQL语句
const CreateTableSQL = `
CREATE TABLE IF NOT EXISTS lynxgraph_info_atom_types (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES system_tenants(id) ON DELETE CASCADE,
	name VARCHAR(255) NOT NULL,
	version VARCHAR(32) NOT NULL DEFAULT 'v1',
	tag_ids UUID[] DEFAULT '{}'::UUID[],
	data_format JSONB NOT NULL,
	created_by VARCHAR(100),
	updated_by VARCHAR(100),
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT uk_lynxgraph_info_atom_types_tenant_name_version UNIQUE(tenant_id, name, version)
);

-- 索引优化
CREATE INDEX IF NOT EXISTS idx_lynxgraph_info_atom_types_tenant ON lynxgraph_info_atom_types(tenant_id);
CREATE INDEX IF NOT EXISTS idx_lynxgraph_info_atom_types_name ON lynxgraph_info_atom_types(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_lynxgraph_info_atom_types_created_at ON lynxgraph_info_atom_types(created_at DESC);

-- 复合索引（常用查询优化）
CREATE INDEX IF NOT EXISTS idx_lynxgraph_info_atom_types_tenant_name ON lynxgraph_info_atom_types(tenant_id, name);

-- 自动更新updated_at字段的触发器
CREATE OR REPLACE FUNCTION update_lynxgraph_info_atom_types_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_lynxgraph_info_atom_types_updated_at ON lynxgraph_info_atom_types;
CREATE TRIGGER trigger_lynxgraph_info_atom_types_updated_at
    BEFORE UPDATE ON lynxgraph_info_atom_types
    FOR EACH ROW
    EXECUTE FUNCTION update_lynxgraph_info_atom_types_updated_at();
`

// CheckTableExistsSQL 检查lynxgraph_info_atom_types表是否存在
const CheckTableExistsSQL = `SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'lynxgraph_info_atom_types'`

// CreateTagAssociationTableSQL 创建信息原子类型标签关联表的PostgreSQL SQL语句
const CreateTagAssociationTableSQL = `
CREATE TABLE IF NOT EXISTS info_atom_type_tags (
    info_atom_type_id UUID NOT NULL REFERENCES lynxgraph_info_atom_types(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES lynxgraph_tag_definitions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (info_atom_type_id, tag_id)
);

-- 索引优化
CREATE INDEX IF NOT EXISTS idx_info_atom_type_tags_type ON info_atom_type_tags(info_atom_type_id);
CREATE INDEX IF NOT EXISTS idx_info_atom_type_tags_tag ON info_atom_type_tags(tag_id);
`
