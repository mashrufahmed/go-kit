package cache

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRedisIntegration(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("set REDIS_ADDR to run Redis integration tests")
	}
	c, err := NewRedis(context.Background(), RedisConfig{Addr: addr, Ping: true})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	ctx := context.Background()
	key := "go-kit:test:cache"
	defer c.Delete(ctx, key)
	if err := c.Set(ctx, key, []byte("value"), time.Minute); err != nil {
		t.Fatal(err)
	}
	value, err := c.Get(ctx, key)
	if err != nil || string(value) != "value" {
		t.Fatalf("get: %q, %v", value, err)
	}
}
