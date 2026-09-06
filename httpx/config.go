package httpx

import "time"

type Config struct {
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	MaxBodySize int64
}

func defaultConfig() Config {
	return Config{
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		MaxBodySize:     1 << 20, // 1 MB
	}
}

func mergeConfig(config Config) Config {
	defaults := defaultConfig()

	if config.ReadTimeout > 0 {
		defaults.ReadTimeout = config.ReadTimeout
	}

	if config.WriteTimeout > 0 {
		defaults.WriteTimeout = config.WriteTimeout
	}

	if config.IdleTimeout > 0 {
		defaults.IdleTimeout = config.IdleTimeout
	}

	if config.ShutdownTimeout > 0 {
		defaults.ShutdownTimeout = config.ShutdownTimeout
	}
	if config.MaxBodySize > 0 {
		defaults.MaxBodySize = config.MaxBodySize
	}

	return defaults
}
