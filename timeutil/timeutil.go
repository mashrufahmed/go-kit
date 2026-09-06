package timeutil

import "time"

func UTC(value time.Time) time.Time       { return value.UTC() }
func Since(value time.Time) time.Duration { return time.Since(value) }
