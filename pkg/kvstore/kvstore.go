package kvstore

import (
	"context"
	"sync"
	"time"
)

type Store[V any] interface {
	Get(ctx context.Context, key string) (V, bool, error)
	Set(ctx context.Context, key string, value V, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type entry[V any] struct {
	value V
	exp   time.Time
}

type Memory[V any] struct {
	mu    sync.Mutex
	items map[string]entry[V]
}

func NewMemory[V any]() *Memory[V] {
	return &Memory[V]{items: map[string]entry[V]{}}
}

func (m *Memory[V]) Get(_ context.Context, key string) (V, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.items[key]
	if !ok || (!e.exp.IsZero() && time.Now().After(e.exp)) {
		if ok {
			delete(m.items, key)
		}
		var zero V
		return zero, false, nil
	}
	return e.value, true, nil
}

func (m *Memory[V]) Set(_ context.Context, key string, value V, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	m.items[key] = entry[V]{value: value, exp: exp}
	return nil
}

func (m *Memory[V]) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	delete(m.items, key)
	m.mu.Unlock()
	return nil
}
