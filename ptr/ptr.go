package ptr

func To[T any](value T) *T { return &value }
func Value[T any](value *T) T {
	if value == nil {
		var zero T
		return zero
	}
	return *value
}
