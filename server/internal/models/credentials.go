package models

import (
	"crypto/sha256"
	"encoding/base64"
	"time"

	i "github.com/luskaner/ageLANServer/server/internal"
)

const credentialsExpiry = time.Minute * 5

func generateSignature() string {
	hash := sha256.Sum256(i.SecureBytes(32))
	return base64.StdEncoding.EncodeToString(hash[:])
}

type credentialKey = string
type credentialValue = string
type Credentials = *BaseSessions[credentialKey, credentialValue]
type Credential = *BaseSession[credentialKey, credentialValue]

func NewCredentials() Credentials {
	return NewBaseSessions[string, string](credentialsExpiry)
}

func CreateCredential(creds Credentials, key *string) Credential {
	return creds.CreateSession(generateSignature, key)
}
