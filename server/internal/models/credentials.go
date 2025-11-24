package models

import (
	"crypto/sha256"
	"encoding/base64"
	"slices"
	"sync"
	"time"

	i "github.com/luskaner/ageLANServer/server/internal"
)

const expiry = time.Minute * 5

type Credentials struct {
	store              *i.SafeMap[string, *Credential]
	sweeperTaskMu      sync.Mutex
	sweeperTaskStarted bool
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
	i.WithRng(func(rand *i.RandReader) {
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
	creds.sweeperTaskMu.Lock()
	defer creds.sweeperTaskMu.Unlock()
	if !creds.sweeperTaskStarted {
		go creds.startSweeper()
		creds.sweeperTaskStarted = true
	}
	return storedCred
}

func (creds *Credentials) nextExpiration() (alreadyExpired []string, nextExpiration time.Duration) {
	var expirationTimes []time.Time
	now := time.Now().UTC()
	for cred := range creds.store.Values() {
		if cred.expiry.Before(now) {
			alreadyExpired = append(alreadyExpired, cred.signature)
		} else {
			expirationTimes = append(expirationTimes, cred.expiry)
		}
	}
	if len(expirationTimes) > 0 {
		slices.SortFunc(expirationTimes, func(a, b time.Time) int {
			switch {
			case a.Before(b):
				return -1
			case a.After(b):
				return 1
			default:
				return 0
			}
		})
		nextExpiration = expirationTimes[0].Sub(now)
	} else {
		nextExpiration = expiry
	}
	return
}

func (creds *Credentials) startSweeper() {
	go func() {
		var alreadyExpired []string
		_, nextExpiration := creds.nextExpiration()
		ticker := time.NewTicker(nextExpiration)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				alreadyExpired, nextExpiration = creds.nextExpiration()
				ticker.Reset(nextExpiration)
				for _, expired := range alreadyExpired {
					creds.store.Delete(expired)
				}
			}
		}
	}()
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
