package internal

import (
	"sync"
)

type SafeSet[K comparable] struct {
	mu   sync.RWMutex
	data map[K]any
}

func NewSafeSet[K comparable]() *SafeSet[K] {
	return &SafeSet[K]{
		data: make(map[K]any),
	}
}

func (sm *SafeSet[K]) Add(key K) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.data[key] = nil
}

func (sm *SafeSet[K]) Has(key K) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	_, ok := sm.data[key]
	return ok
}

func (sm *SafeSet[K]) Delete(key K) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.data, key)
}

func (sm *SafeSet[K]) Len() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.data)
}

func (sm *SafeSet[K]) Iter() <-chan K {
	ch := make(chan K)
	go func() {
		sm.mu.RLock()
		defer sm.mu.RUnlock()
		for k := range sm.data {
			ch <- k
		}
		close(ch)
	}()
	return ch
}
