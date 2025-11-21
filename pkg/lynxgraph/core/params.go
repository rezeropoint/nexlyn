package core

type GraphList struct {
	Page     int64    `form:"page,optional,default=1"`      // 页码，从1开始
	PageSize int64    `form:"pageSize,optional,default=10"` // 每页大小
	TenantId string   `form:"tenantId,optional"`            // 租户ID筛选
	OrgIDs   []string `form:"-"`                            // 组织ID列表（用于权限过滤，不从前端接收）
	Name     string   `form:"name,optional"`                // 名称模糊查询
	Tags     []string `form:"tags,optional"`                // 标签筛选（数组）
	Enable   *bool    `form:"enable,optional"`              // 启用状态筛选
}

type InfoAtomList struct {
	Page     int64    `form:"page,optional,default=1"`      // 页码，从1开始
	PageSize int64    `form:"pageSize,optional,default=10"` // 每页大小
	TenantId string   `form:"tenantId,optional"`            // 租户ID筛选
	Name     string   `form:"name,optional"`                // 名称模糊查询
	Tags     []string `form:"tags,optional"`                // 标签筛选（数组）
}
