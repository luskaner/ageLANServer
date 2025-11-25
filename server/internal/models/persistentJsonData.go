package models

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sync"

	"github.com/luskaner/ageLANServer/common/fileLock"
)

type Equalable[T any] interface {
	Equals(other *T) bool
}

func decodeData[T any](f *os.File) (data T, err error) {
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&data)
	return
}

func openFile(p string) (existed bool, f *os.File, err error) {
	if f, err = os.OpenFile(p, os.O_RDWR, 0644); err == nil {
		existed = true
		return
	} else if errors.Is(err, fs.ErrNotExist) {
		f, err = os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0644)
	}
	return
}

type PersistentJsonROData[T any] struct {
	data *T
}

func (d *PersistentJsonROData[T]) Data() T {
	return *d.data
}

func NewPersistentJsonROData[T any](path string) (userData *PersistentJsonROData[T], err error) {
	var existed bool
	var f *os.File
	existed, f, err = openFile(path)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	var data T
	if existed {
		data, err = decodeData[T](f)
		if err != nil {
			return
		}
	}
	userData = &PersistentJsonROData[T]{&data}
	return
}

type PersistentJsonData[T Equalable[T]] struct {
	*PersistentJsonROData[T]
	lock     *sync.RWMutex
	fileLock *fileLock.Lock
}

func (d *PersistentJsonData[T]) Update(data *T) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if (*data).Equals(d.data) {
		return
	}
	if _, err = d.fileLock.File.Seek(0, 0); err != nil {
		return
	}
	if err = d.fileLock.File.Truncate(0); err != nil {
		return
	}
	encoder := json.NewEncoder(d.fileLock.File)
	if err = encoder.Encode(data); err != nil {
		return
	}
	_ = d.fileLock.File.Sync()
	*d.data = *data
	return
}

func (d *PersistentJsonData[T]) Data() T {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.PersistentJsonROData.Data()
}

func (d *PersistentJsonData[T]) Close() {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.fileLock != nil {
		_ = d.fileLock.Unlock()
		d.fileLock = nil
		d.data = nil
	}
}

func NewPersistentJsonData[T Equalable[T]](path string) (persistentData *PersistentJsonData[T], err error) {
	var existed bool
	var f *os.File
	existed, f, err = openFile(path)
	if err != nil {
		return
	}
	lock := fileLock.Lock{}
	if err = lock.Lock(f); err != nil {
		_ = f.Close()
		return
	}
	var data T
	if existed {
		data, err = decodeData[T](f)
		if err != nil {
			_ = lock.Unlock()
			return
		}
		if _, err = f.Seek(0, 0); err != nil {
			_ = lock.Unlock()
			return
		}
	}
	persistentData = &PersistentJsonData[T]{
		&PersistentJsonROData[T]{&data},
		&sync.RWMutex{},
		&lock,
	}
	return
}
