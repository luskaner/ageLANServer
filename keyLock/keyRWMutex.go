package keyLock

import "sync"

type KeyRWMutex struct {
	mu      sync.RWMutex
	mutexes map[interface{}]*sync.RWMutex
}

func NewKeyRWMutex() *KeyRWMutex {
	return &KeyRWMutex{
		mutexes: make(map[interface{}]*sync.RWMutex),
	}
}

func (kl *KeyRWMutex) Lock(key interface{}) {
	kl.mu.Lock()
	lock, ok := kl.mutexes[key]
	if !ok {
		lock = &sync.RWMutex{}
		kl.mutexes[key] = lock
	}
	kl.mu.Unlock()

	lock.Lock()
}

func (kl *KeyRWMutex) RLock(key interface{}) {
	kl.mu.RLock()
	lock, ok := kl.mutexes[key]
	kl.mu.RUnlock()

	if ok {
		lock.RLock()
	}
}

func (kl *KeyRWMutex) Unlock(key interface{}) {
	kl.mu.RLock()
	lock, ok := kl.mutexes[key]
	kl.mu.RUnlock()

	if ok {
		lock.Unlock()
	}
}

func (kl *KeyRWMutex) RUnlock(key interface{}) {
	kl.mu.RLock()
	lock, ok := kl.mutexes[key]
	kl.mu.RUnlock()

	if ok {
		lock.RUnlock()
	}
}
