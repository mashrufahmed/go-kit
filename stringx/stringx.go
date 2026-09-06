package stringx

import "strings"

func IsBlank(value string) bool { return strings.TrimSpace(value) == "" }
func Truncate(value string, n int) string {
	r := []rune(value)
	if n < 0 {
		n = 0
	}
	if len(r) <= n {
		return value
	}
	return string(r[:n])
}
