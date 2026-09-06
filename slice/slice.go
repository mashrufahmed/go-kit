package slice

func Map[T, U any](values []T, fn func(T) U) []U {
	out := make([]U, len(values))
	for i, value := range values {
		out[i] = fn(value)
	}
	return out
}
func Filter[T any](values []T, fn func(T) bool) []T {
	out := make([]T, 0, len(values))
	for _, value := range values {
		if fn(value) {
			out = append(out, value)
		}
	}
	return out
}
func Contains[T comparable](values []T, value T) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
