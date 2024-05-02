package extra

import (
	"aoe2DELanServer/rng"
	"crypto/sha256"
	"encoding/base64"
	"sync"
	"time"
)

type Info struct {
	signature string
	key       string
	expiry    time.Time
}

var store sync.Map

func (info *Info) Delete() {
	store.Delete(info.signature)
}

func generateSignature() string {
	b := make([]byte, 32)
	for {
		rng.Rng.Read(b)
		hash := sha256.Sum256(b)
		sig := base64.StdEncoding.EncodeToString(hash[:])
		if _, exists := Get(sig); !exists {
			return sig
		}
	}
}

func Create(key string) *Info {
	removeExpired()
	info := &Info{
		key:       key,
		signature: generateSignature(),
		expiry:    time.Now().UTC().Add(time.Minute * 5),
	}
	store.Store(info.signature, info)
	return info
}

func Get(signature string) (*Info, bool) {
	value, exists := store.Load(signature)
	if exists {
		info := value.(*Info)
		if info.Expired() {
			info.Delete()
			exists = false
			info = nil
		}
		return info, exists
	}
	return nil, false
}

func (info *Info) Expired() bool {
	return time.Now().UTC().After(info.expiry)
}

func (info *Info) GetExpiry() time.Time {
	return info.expiry
}

func (info *Info) GetSignature() string {
	return info.signature
}

func (info *Info) GetKey() string {
	return info.key
}

func removeExpired() {
	store.Range(func(_, value interface{}) bool {
		info := value.(*Info)
		if info.Expired() {
			info.Delete()
		}
		return true
	})
}
