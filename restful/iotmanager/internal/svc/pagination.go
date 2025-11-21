package svc

// NormalizePagination 规范化分页参数
// current: 当前页码（从1开始）
// pageSize: 每页大小
// 返回: normalizedCurrent, normalizedPageSize, offset
func NormalizePagination(current, pageSize int64) (normalizedCurrent, normalizedPageSize, offset int64) {
	// 确保页码至少为1
	if current <= 0 {
		current = 1
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
	offset = (current - 1) * pageSize

	return current, pageSize, offset
}
