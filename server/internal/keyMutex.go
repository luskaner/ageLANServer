package internal

import "sync"

type KeyMutex struct {
	mu      sync.Mutex
	mutexes map[interface{}]*sync.Mutex
}

func NewKeyMutex() *KeyMutex {
	return &KeyMutex{
		mutexes: make(map[interface{}]*sync.Mutex),
	}
}

func (kl *KeyMutex) Lock(key interface{}) {
	kl.mu.Lock()
	lock, ok := kl.mutexes[key]
	if !ok {
		lock = &sync.Mutex{}
		kl.mutexes[key] = lock
	}

	kl.mu.Unlock()
	lock.Lock()
}

func (kl *KeyMutex) Unlock(key interface{}) {
	kl.mu.Lock()
	lock, ok := kl.mutexes[key]
	kl.mu.Unlock()

	if ok {
		lock.Unlock()
	}
}
