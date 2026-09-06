package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig configures a Redis-backed cache.
type RedisConfig struct {
	Addr         string
	Username     string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Ping         bool
}

// Redis is a Cache implementation backed by Redis.
type Redis struct{ client *redis.Client }

// NewRedis creates a Redis cache. When Ping is true, it verifies connectivity before returning.
func NewRedis(ctx context.Context, cfg RedisConfig) (*Redis, error) {
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:6379"
	}
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})
	r := &Redis{client: client}
	if cfg.Ping {
		if err := client.Ping(ctx).Err(); err != nil {
			_ = client.Close()
			return nil, err
		}
	}
	return r, nil
}

func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	return value, err
}

func (r *Redis) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if ttl == -2*time.Second {
		return 0, ErrNotFound
	}
	return ttl, nil
}

func (r *Redis) JSONSet(ctx context.Context, key string, value any, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, b, ttl)
}

func (r *Redis) JSONGet(ctx context.Context, key string, target any) error {
	b, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

func (r *Redis) Close() error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Close()
}

func (r *Redis) Client() *redis.Client { return r.client }
