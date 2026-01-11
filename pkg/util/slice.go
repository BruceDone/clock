package util

// Contains 检查切片是否包含元素
func Contains[T comparable](slice []T, elem T) bool {
	for _, v := range slice {
		if v == elem {
			return true
		}
	}
	return false
}

// ContainsInt 检查int切片是否包含元素
func ContainsInt(slice []int, elem int) bool {
	return Contains(slice, elem)
}

// ContainsString 检查string切片是否包含元素
func ContainsString(slice []string, elem string) bool {
	return Contains(slice, elem)
}

// Filter 过滤切片
func Filter[T any](slice []T, fn func(T) bool) []T {
	result := make([]T, 0)
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Map 映射切片
func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}
