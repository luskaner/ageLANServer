package models

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"sync"

	"github.com/luskaner/ageLANServer/common/fileLock"
	"github.com/luskaner/ageLANServer/server/internal"
)

type UpgradableData[T any] interface {
	ObjectOfVersion(version uint32) any
	CurrentVersion() uint32
	UpgradeToNextVersion(oldVersion uint32, oldObject any) T
}

type InitialUpgradableData[T any] struct {
}

func (i *InitialUpgradableData[T]) ObjectOfVersion(_ uint32) any {
	panic("should not have been called")
}

func (i *InitialUpgradableData[T]) CurrentVersion() uint32 {
	return 0
}

func (i *InitialUpgradableData[T]) UpgradeToNextVersion(_ uint32, _ any) T {
	panic("should not have been called")
}

type DefaultUpgradableData[T any] interface {
	UpgradableData[T]
	Default() T
}

func upgrade[T any](file *PersistentFile, version uint32, upgrader UpgradableData[T]) (err error, upgraded bool, data T) {
	versionsToUpgrade := upgrader.CurrentVersion() - version
	if versionsToUpgrade == 0 {
		return
	}
	currentData := upgrader.ObjectOfVersion(version)
	if err = readPersistentData(file, &currentData); err != nil {
		return
	}
	for versionsUpgraded := uint32(0); versionsUpgraded < versionsToUpgrade; versionsUpgraded++ {
		currentData = upgrader.UpgradeToNextVersion(version+versionsUpgraded, currentData)
	}
	data = currentData.(T)
	upgraded = true
	return
}

type InitialUpgradableDefaultData[T any] struct {
	InitialUpgradableData[T]
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

type PersistentFile struct {
	lock     *sync.Mutex
	fileLock *fileLock.Lock
	existed  bool
}

func (d *PersistentFile) Existed() bool {
	return d.existed
}

func (d *PersistentFile) unsafeSeekTop() (err error) {
	_, err = d.fileLock.File.Seek(0, 0)
	return
}

func (d *PersistentFile) WithWriter(fn func(writer io.Writer) error) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if err = d.unsafeSeekTop(); err != nil {
		return
	}
	if err = d.fileLock.File.Truncate(0); err != nil {
		return
	}
	err = fn(d.fileLock.File)
	_ = d.fileLock.File.Sync()
	return err
}

func (d *PersistentFile) WithReader(fn func(writer io.Reader) error) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if err = d.unsafeSeekTop(); err != nil {
		return
	}
	return fn(d.fileLock.File)
}

func (d *PersistentFile) Close() (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.fileLock.Unlock()
}

func NewPersistentData(path string) (data *PersistentFile, err error) {
	var f *os.File
	var existed bool
	existed, f, err = openFile(path)
	if err != nil {
		return
	}
	lock := fileLock.Lock{}
	if err = lock.Lock(f); err != nil {
		_ = f.Close()
		return
	}
	data = &PersistentFile{
		&sync.Mutex{},
		&lock,
		existed,
	}
	return
}

type jsonMetadata struct {
	Version uint32 `json:"version"`
}

type jsonDataWithMetadata[T any] struct {
	Metadata jsonMetadata `json:"metadata"`
	Data     T            `json:"data"`
}

type jsonAnyDataWithMetadata struct {
	Metadata jsonMetadata `json:"xboxSessionId"`
	Data     any          `json:"data"`
}

func readPersistentData(persistentFile *PersistentFile, data any) (err error) {
	return persistentFile.WithReader(func(reader io.Reader) error {
		return json.NewDecoder(reader).Decode(data)
	})
}

type PersistentStringJsonMapRaw = internal.SafeMap[string, jsonDataWithMetadata[json.RawMessage]]
type PersistentStringJsonMapDataMap = internal.SafeMap[string, jsonAnyDataWithMetadata]

type PersistentStringJsonMap struct {
	file            *PersistentFile
	cachedRawData   jsonDataWithMetadata[*PersistentStringJsonMapRaw]
	currentData     jsonDataWithMetadata[*PersistentStringJsonMapDataMap]
	currentDataLock *internal.KeyRWMutex[string]
}

func NewPersistentStringMap(path string, upgrader UpgradableData[*PersistentStringJsonMapRaw]) (persistentMap *PersistentStringJsonMap, err error) {
	var file *PersistentFile
	file, err = NewPersistentData(path)
	if err != nil {
		return
	}
	var initialRawData *PersistentStringJsonMapRaw
	if file.Existed() {
		var metadata jsonMetadata
		if err = readPersistentData(file, &metadata); err != nil {
			return
		}
		if metadata.Version > upgrader.CurrentVersion() {
			_ = file.fileLock.Unlock()
			err = errors.New("data version is newer than current version")
			return
		} else if localErr, upgraded, data := upgrade(file, metadata.Version, upgrader); localErr == nil && upgraded {
			initialRawData = data
		} else if localErr != nil {
			_ = file.fileLock.Unlock()
			err = localErr
			return
		} else {
			if err = readPersistentData(file, &initialRawData); err != nil {
				return
			}
		}
	} else {
		initialRawData = internal.NewSafeMap[string, jsonDataWithMetadata[json.RawMessage]]()
	}
	persistentMap = &PersistentStringJsonMap{
		file,
		jsonDataWithMetadata[*PersistentStringJsonMapRaw]{
			Metadata: jsonMetadata{Version: upgrader.CurrentVersion()},
			Data:     initialRawData,
		},
		jsonDataWithMetadata[*PersistentStringJsonMapDataMap]{
			Metadata: jsonMetadata{Version: upgrader.CurrentVersion()},
			Data:     internal.NewSafeMap[string, jsonAnyDataWithMetadata](),
		},
		internal.NewKeyRWMutex[string](),
	}
	if !file.Existed() {
		err = file.WithWriter(func(writer io.Writer) error {
			return json.NewEncoder(writer).Encode(persistentMap.cachedRawData)
		})
	}
	return
}

func psmjSet[T any](p *PersistentStringJsonMap, key string, value DefaultUpgradableData[T]) (err error) {
	p.currentDataLock.Lock(key)
	defer p.currentDataLock.Unlock(key)
	var data T
	var saveToCache bool
	if currentVal, ok := p.cachedRawData.Data.Load(key); !ok {
		data = value.Default()
		saveToCache = true
	} else if valueCurrentVersion := value.CurrentVersion(); currentVal.Metadata.Version > valueCurrentVersion {
		return errors.New("data version is newer than current version")
	} else if localErr, upgraded, d := upgrade(p.file, currentVal.Metadata.Version, value); localErr == nil && upgraded {
		data = d
		saveToCache = true
	} else if localErr != nil {
		err = localErr
		return
	} else if err = json.Unmarshal(currentVal.Data, &data); err != nil {
		return
	}
	finalData := jsonAnyDataWithMetadata{
		jsonMetadata{value.CurrentVersion()},
		data,
	}
	if _, exists := (*p.currentData.Data).Store(key, finalData, func(_ jsonAnyDataWithMetadata) bool {
		return false
	}); exists {
		return errors.New("key already exists")
	}
	if saveToCache {
		cacheData := jsonDataWithMetadata[json.RawMessage]{
			Metadata: jsonMetadata{value.CurrentVersion()},
		}
		if cacheData.Data, err = json.Marshal(data); err != nil {
			return
		}
		p.cachedRawData.Data.Store(key, cacheData, func(stored jsonDataWithMetadata[json.RawMessage]) bool {
			return true
		})
		return p.file.WithWriter(func(writer io.Writer) error {
			return json.NewEncoder(writer).Encode(p.cachedRawData)
		})
	}
	return nil
}

func psmjFn[T any](p *PersistentStringJsonMap, key string, fn func(data T) error) (fullData jsonAnyDataWithMetadata, err error) {
	var ok bool
	fullData, ok = (*p.currentData.Data).Load(key)
	if !ok {
		err = errors.New("key does not exist")
		return
	}
	err = fn(fullData.Data.(T))
	return
}

func psmjWithReadOnly[T any](p *PersistentStringJsonMap, key string, fn func(data T) error) error {
	p.currentDataLock.RLock(key)
	defer p.currentDataLock.RUnlock(key)
	_, err := psmjFn(p, key, fn)
	return err
}

func psmjWithReadWrite[T any](p *PersistentStringJsonMap, key string, fn func(data T) error) error {
	p.currentDataLock.Lock(key)
	defer p.currentDataLock.Unlock(key)
	var fullData jsonAnyDataWithMetadata
	var err error
	if fullData, err = psmjFn(p, key, fn); err != nil {
		return err
	}
	var dataBytes []byte
	if dataBytes, err = json.Marshal(fullData.Data); err == nil {
		_, _ = p.cachedRawData.Data.Store(key, jsonDataWithMetadata[json.RawMessage]{
			Metadata: fullData.Metadata,
			Data:     dataBytes,
		}, func(_ jsonDataWithMetadata[json.RawMessage]) bool {
			return true
		})
		return p.file.WithWriter(func(writer io.Writer) error {
			return json.NewEncoder(writer).Encode(p.cachedRawData)
		})
	} else {
		return err
	}
}

type PersistentJsonData[T any] struct {
	persistentMap *PersistentStringJsonMap
	key           string
}

func (p *PersistentJsonData[T]) WithReadOnly(fn func(data T) error) error {
	return psmjWithReadOnly[T](p.persistentMap, p.key, fn)
}

func (p *PersistentJsonData[T]) WithReadWrite(fn func(data T) error) error {
	return psmjWithReadWrite[T](p.persistentMap, p.key, fn)
}

func (p *PersistentJsonData[T]) MarshalJSON() (data []byte, err error) {
	err = p.WithReadOnly(func(d T) error {
		data, err = json.Marshal(d)
		return err
	})
	return
}

func NewPersistentJsonData[T any](persistentMap *PersistentStringJsonMap, key string, upgrader DefaultUpgradableData[T]) (data *PersistentJsonData[T], err error) {
	if err = psmjSet(persistentMap, key, upgrader); err != nil {
		return
	}
	data = &PersistentJsonData[T]{
		persistentMap: persistentMap,
		key:           key,
	}
	return
}
