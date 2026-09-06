package maps

func Keys[K comparable, V any](value map[K]V) []K {
	out := make([]K, 0, len(value))
	for key := range value {
		out = append(out, key)
	}
	return out
}
func Values[K comparable, V any](value map[K]V) []V {
	out := make([]V, 0, len(value))
	for _, item := range value {
		out = append(out, item)
	}
	return out
}
