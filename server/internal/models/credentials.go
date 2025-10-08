package models

import (
	"crypto/sha256"
	"encoding/base64"
	"math/rand/v2"
	"time"

	i "github.com/luskaner/ageLANServer/server/internal"
)

const expiry = time.Minute * 5

type Credentials struct {
	store *i.SafeMap[string, *Credential]
}

type Credential struct {
	signature string
	key       string
	expiry    time.Time
}

func (creds *Credentials) Initialize() {
	creds.store = i.NewSafeMap[string, *Credential]()
}

func (creds *Credentials) generateSignature() string {
	b := make([]byte, 32)
	i.WithRng(func(rand *rand.Rand) {
		for j := 0; j < len(b); j++ {
			b[j] = byte(rand.UintN(256))
		}
	})
	hash := sha256.Sum256(b)
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (creds *Credentials) CreateCredentials(key string) *Credential {
	var storedCred *Credential
	for exists := true; exists; {
		info := &Credential{
			key:       key,
			signature: creds.generateSignature(),
			expiry:    time.Now().UTC().Add(expiry),
		}
		storedCred, exists = creds.store.Store(info.signature, info, func(_ *Credential) bool {
			return false
		})
	}
	time.AfterFunc(expiry, func() {
		creds.store.Delete(storedCred.signature)
	})
	return storedCred
}

func (creds *Credentials) GetCredentials(signature string) (*Credential, bool) {
	return creds.store.Load(signature)
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
