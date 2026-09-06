package cache

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("cache: key not found")

type Cache interface {
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, []byte, time.Duration) error
	Delete(context.Context, string) error
	Exists(context.Context, string) (bool, error)
	TTL(context.Context, string) (time.Duration, error)
	Close() error
}
type item struct {
	value   []byte
	expires time.Time
}
type Memory struct {
	mu    sync.RWMutex
	items map[string]item
}

func NewMemory() *Memory { return &Memory{items: make(map[string]item)} }
func (m *Memory) Get(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	v, ok := m.items[key]
	m.mu.RUnlock()
	if !ok || expired(v) {
		if ok {
			_ = m.Delete(ctx, key)
		}
		return nil, ErrNotFound
	}
	return append([]byte(nil), v.value...), nil
}
func (m *Memory) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	expiry := time.Time{}
	if ttl > 0 {
		expiry = time.Now().Add(ttl)
	}
	m.mu.Lock()
	m.items[key] = item{append([]byte(nil), value...), expiry}
	m.mu.Unlock()
	return nil
}
func (m *Memory) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	delete(m.items, key)
	m.mu.Unlock()
	return nil
}
func (m *Memory) Exists(ctx context.Context, key string) (bool, error) {
	_, err := m.Get(ctx, key)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}
func (m *Memory) TTL(ctx context.Context, key string) (time.Duration, error) {
	v, err := m.Get(ctx, key)
	_ = v
	if err != nil {
		return 0, err
	}
	m.mu.RLock()
	expiry := m.items[key].expires
	m.mu.RUnlock()
	if expiry.IsZero() {
		return -1, nil
	}
	return time.Until(expiry), nil
}
func (m *Memory) JSONSet(ctx context.Context, key string, value any, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return m.Set(ctx, key, b, ttl)
}
func (m *Memory) JSONGet(ctx context.Context, key string, target any) error {
	b, err := m.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}
func (m *Memory) Close() error { m.mu.Lock(); clear(m.items); m.mu.Unlock(); return nil }
func expired(v item) bool      { return !v.expires.IsZero() && time.Now().After(v.expires) }
