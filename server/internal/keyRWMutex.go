package internal

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
	ok, lock := kl.rlock(key)
	if !ok {
		lock = &sync.RWMutex{}
		func() {
			kl.mu.Lock()
			defer kl.mu.Unlock()
			kl.mutexes[key] = lock
		}()
	}
	lock.Lock()
}

func (kl *KeyRWMutex) rlock(key interface{}) (ok bool, lock *sync.RWMutex) {
	kl.mu.RLock()
	defer kl.mu.RUnlock()
	lock, ok = kl.mutexes[key]
	return
}

func (kl *KeyRWMutex) RLock(key interface{}) {
	ok, lock := kl.rlock(key)

	if ok {
		lock.RLock()
	}
}

func (kl *KeyRWMutex) Unlock(key interface{}) {
	ok, lock := kl.rlock(key)

	if ok {
		lock.Unlock()
	}
}

func (kl *KeyRWMutex) RUnlock(key interface{}) {
	ok, lock := kl.rlock(key)

	if ok {
		lock.RUnlock()
	}
}
