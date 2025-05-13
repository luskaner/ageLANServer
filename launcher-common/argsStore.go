package launcher_common

import (
	mapset "github.com/deckarep/golang-set/v2"
	"io"
	"os"
	"strings"
	"sync"
)

const argsStoreSep = "|"

func argsStoreByteToStringSlice(s []byte) []string {
	return strings.Split(string(s), argsStoreSep)
}

type ArgsStore struct {
	filePath string
	mutex    sync.RWMutex
}

func NewArgsStore(filePath string) *ArgsStore {
	return &ArgsStore{
		filePath: filePath,
	}
}

func (s *ArgsStore) Load() (err error, flags []string) {
	var content []byte
	s.mutex.RLock()
	func() {
		defer s.mutex.RUnlock()
		content, err = os.ReadFile(s.filePath)
	}()
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}
	flags = argsStoreByteToStringSlice(content)
	return
}

func (s *ArgsStore) Store(flags []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	f, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	var content []byte
	content, err = io.ReadAll(f)
	if err != nil {
		return err
	}
	flagsToSave := argsStoreByteToStringSlice(content)
	existingFlags := mapset.NewSet[string](flagsToSave...)
	for _, flag := range flags {
		if !existingFlags.ContainsOne(flag) {
			flagsToSave = append(flagsToSave, flag)
		}
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	_, err = f.WriteString(strings.Join(flagsToSave, argsStoreSep))
	return err
}

func (s *ArgsStore) Delete() error {
	s.mutex.Lock()
	var err error
	func() {
		defer s.mutex.Unlock()
		err = os.Remove(s.filePath)
	}()
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
