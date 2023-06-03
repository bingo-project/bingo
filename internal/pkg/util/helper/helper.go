package helper

const defaultLimitValue = 20

// DefaultLimit 设置默认查询记录数.
func DefaultLimit(limit int) int {
	if limit == 0 {
		limit = defaultLimitValue
	}

	return limit
}
