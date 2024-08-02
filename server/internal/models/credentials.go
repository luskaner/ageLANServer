package models

import (
	"crypto/sha256"
	"encoding/base64"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"sync"
	"time"
)

type Credentials struct {
	signature string
	key       string
	expiry    time.Time
}

var store sync.Map

func (info *Credentials) Delete() {
	store.Delete(info.signature)
}

func generateSignature() string {
	b := make([]byte, 32)
	for {
		i.Rng.Read(b)
		hash := sha256.Sum256(b)
		sig := base64.StdEncoding.EncodeToString(hash[:])
		if _, exists := GetCredentials(sig); !exists {
			return sig
		}
	}
}

func CreateCredentials(key string) *Credentials {
	removeCredentialsExpired()
	info := &Credentials{
		key:       key,
		signature: generateSignature(),
		expiry:    time.Now().UTC().Add(time.Minute * 5),
	}
	store.Store(info.signature, info)
	return info
}

func GetCredentials(signature string) (*Credentials, bool) {
	value, exists := store.Load(signature)
	if exists {
		info := value.(*Credentials)
		if info.Expired() {
			info.Delete()
			exists = false
			info = nil
		}
		return info, exists
	}
	return nil, false
}

func (info *Credentials) Expired() bool {
	return time.Now().UTC().After(info.expiry)
}

func (info *Credentials) GetExpiry() time.Time {
	return info.expiry
}

func (info *Credentials) GetSignature() string {
	return info.signature
}

func (info *Credentials) GetKey() string {
	return info.key
}

func removeCredentialsExpired() {
	store.Range(func(_, value interface{}) bool {
		info := value.(*Credentials)
		if info.Expired() {
			info.Delete()
		}
		return true
	})
}
