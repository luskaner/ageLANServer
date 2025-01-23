package internal

import (
	"iter"
	"sync"
)

type SafeMap[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		data: make(map[K]V),
	}
}

func (sm *SafeMap[K, V]) Store(key K, value V) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.data[key] = value
}

func (sm *SafeMap[K, V]) Load(key K) (V, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	value, ok := sm.data[key]
	return value, ok
}

func (sm *SafeMap[K, V]) Delete(key K) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.data, key)
}

func (sm *SafeMap[K, V]) Len() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.data)
}

func (sm *SafeMap[K, V]) Iter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		sm.mu.RLock()
		defer sm.mu.RUnlock()

		for k, v := range sm.data {
			if !yield(k, v) {
				return
			}
		}
	}
}
