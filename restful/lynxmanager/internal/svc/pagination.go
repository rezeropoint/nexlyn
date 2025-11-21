package svc

// NormalizePagination 规范化分页参数
// page: 页码（从1开始）
// pageSize: 每页大小
// 返回: normalizedPage, normalizedPageSize, offset
func NormalizePagination(page, pageSize int) (normalizedPage, normalizedPageSize, offset int) {
	// 确保页码至少为1
	if page <= 0 {
		page = 1
	}

	// 确保页大小至少为1，默认为10
	if pageSize <= 0 {
		pageSize = 10
	}

	// 限制最大页大小为100
	if pageSize > 100 {
		pageSize = 100
	}

	// 计算偏移量
	offset = (page - 1) * pageSize

	return page, pageSize, offset
}
