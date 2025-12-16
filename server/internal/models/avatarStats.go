package models

import (
	"encoding/json"
	"time"

	i "github.com/luskaner/ageLANServer/server/internal"
)

type AvatarStat struct {
	Id int32 `json:"-"`
	// Read Value with AvatarStats.WithReadLock
	Value int64
	// Read Metadata with AvatarStats.WithReadLock
	Metadata json.RawMessage `json:",omitempty"`
	// Read Metadata with AvatarStats.WithReadLock
	LastUpdated time.Time
}

func NewAvatarStat(id int32, value int64) AvatarStat {
	return AvatarStat{
		Id:          id,
		Value:       value,
		LastUpdated: time.Now().UTC(),
	}
}

// UnsafeSetValue requires using AvatarStats.WithWriteLock
func (as *AvatarStat) UnsafeSetValue(value int64) {
	as.Value = value
	as.LastUpdated = time.Now().UTC()
}

// UnsafeEncode requires using AvatarStats.WithReadLock
func (as *AvatarStat) UnsafeEncode(profileId int32) i.A {
	return i.A{
		as.Id,
		profileId,
		as.Value,
		as.Metadata,
		as.LastUpdated.Unix(),
	}
}

type AvatarStats struct {
	values *i.SafeMap[int32, AvatarStat]
	locks  *i.KeyRWMutex[int32]
}

func (as *AvatarStats) MarshalJSON() ([]byte, error) {
	data := make(map[int32]AvatarStat, as.values.Len())
	for stat := range as.values.Values() {
		data[stat.Id] = stat
	}
	return json.Marshal(data)
}

func (as *AvatarStats) UnmarshalJSON(b []byte) error {
	var data map[int32]AvatarStat
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	as.values = i.NewSafeMap[int32, AvatarStat]()
	as.locks = i.NewKeyRWMutex[int32]()
	for id, stat := range data {
		stat.Id = id
		as.values.Store(stat.Id, stat, func(stored AvatarStat) bool {
			return true
		})
	}
	return nil
}

func (as *AvatarStats) GetStat(id int32) (AvatarStat, bool) {
	return as.values.Load(id)
}

// AddStat Must be ensured that the stat does not already exist
func (as *AvatarStats) AddStat(avatarStat AvatarStat) {
	as.values.Store(avatarStat.Id, avatarStat, func(stored AvatarStat) bool {
		return false
	})
}

// Encode must only be called in the same goroutine it is created in
func (as *AvatarStats) Encode(profileId int32) i.A {
	result := i.A{}
	for val := range as.values.Values() {
		result = append(result, val.UnsafeEncode(profileId))
	}
	return result
}

func (as *AvatarStats) WithReadLock(id int32, action func()) {
	as.locks.RLock(id)
	defer as.locks.RUnlock(id)
	action()
}

func (as *AvatarStats) WithWriteLock(id int32, action func()) {
	as.locks.Lock(id)
	defer as.locks.Unlock(id)
	action()
}
