package models

import (
	"crypto/sha256"
	"encoding/base64"
	mapset "github.com/deckarep/golang-set/v2"
	i "github.com/luskaner/ageLANServer/server/internal"
	"sync"
	"time"
)

type Credentials struct {
	store    *i.SafeMap[string, *Credential]
	hashLock sync.Mutex
}

type Credential struct {
	signature string
	key       string
	expiry    time.Time
}

func (creds *Credentials) Initialize() {
	creds.store = i.NewSafeMap[string, *Credential]()
}

func (creds *Credentials) Delete(signature string) {
	creds.store.Delete(signature)
}

func (creds *Credentials) generateSignature() string {
	b := make([]byte, 32)
	for {
		func() {
			i.RngLock.Lock()
			defer i.RngLock.Unlock()
			for j := 0; j < 32; j++ {
				b[j] = byte(i.Rng.UintN(256))
			}
		}()
		var hash [32]byte
		func() {
			creds.hashLock.Lock()
			defer creds.hashLock.Unlock()
			hash = sha256.Sum256(b)
		}()
		sig := base64.StdEncoding.EncodeToString(hash[:])
		if _, exists := creds.GetCredentials(sig); !exists {
			return sig
		}
	}
}

func (creds *Credentials) CreateCredentials(key string) *Credential {
	creds.removeCredentialsExpired()
	info := &Credential{
		key:       key,
		signature: creds.generateSignature(),
		expiry:    time.Now().UTC().Add(time.Minute * 5),
	}
	creds.store.Store(info.signature, info)
	return info
}

func (creds *Credentials) GetCredentials(signature string) (*Credential, bool) {
	cred, exists := creds.store.Load(signature)
	if exists && cred.Expired() {
		creds.Delete(cred.signature)
		exists = false
		cred = nil
	}
	return cred, exists
}

func (cred *Credential) Expired() bool {
	return time.Now().UTC().After(cred.expiry)
}

func (cred *Credential) GetExpiry() time.Time {
	return cred.expiry
}

func (cred *Credential) GetSignature() string {
	return cred.signature
}

func (cred *Credential) GetKey() string {
	return cred.key
}

func (creds *Credentials) removeCredentialsExpired() {
	signaturesToRemove := mapset.NewThreadUnsafeSet[string]()
	for _, credKey := range creds.store.IterKeys() {
		credValue, ok := creds.store.Load(credKey)
		if ok && credValue.Expired() {
			signaturesToRemove.Add(credKey)
		}
	}
	for sig := range signaturesToRemove.Iter() {
		creds.Delete(sig)
	}
}
