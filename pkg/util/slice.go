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
