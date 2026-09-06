package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryTTL(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()
	if err := c.Set(ctx, "k", []byte("v"), time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Millisecond)
	if _, err := c.Get(ctx, "k"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
