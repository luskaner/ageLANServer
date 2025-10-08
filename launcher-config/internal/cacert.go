package internal

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	launcherCommonCert "github.com/luskaner/ageLANServer/launcher-common/cert"
)

type CACert struct {
	gamePath string
}

func NewCACert(gameId string, gamePath string) *CACert {
	if gameId == common.GameAoE2 {
		gamePath = filepath.Join(gamePath, "certificates")
	}
	return &CACert{gamePath: gamePath}
}

func (c *CACert) name() string {
	return "cacert.pem"
}

func (c *CACert) originalPath() string {
	return filepath.Join(c.gamePath, c.name())
}

func (c *CACert) tmpPath() string {
	return filepath.Join(c.gamePath, c.name()+".lan")
}

func (c *CACert) backupPath() string {
	return filepath.Join(c.gamePath, c.name()+".bak")
}

func (c *CACert) Backup() (err error) {
	originalPath := c.originalPath()
	if _, err = os.Stat(originalPath); err != nil {
		return
	}
	backupPath := c.backupPath()
	if _, err = os.Stat(backupPath); err == nil {
		return
	}
	originalFile, err := os.Open(originalPath)
	if err != nil {
		return
	}
	defer func(originalFile *os.File) {
		_ = originalFile.Close()
	}(originalFile)

	backupFile, err := os.Create(backupPath)
	if err != nil {
		return
	}
	defer func(backupFile *os.File) {
		_ = backupFile.Close()
	}(backupFile)

	_, err = io.Copy(backupFile, originalFile)
	if err != nil {
		_ = backupFile.Close()
		_ = os.Remove(backupPath)
		return
	}

	_ = backupFile.Sync()
	return
}

func readCertsFromFile(filePath string) (keys []string, keyToIndex map[string]int, values []*x509.Certificate, err error) {
	var pemData []byte
	pemData, err = os.ReadFile(filePath)
	if err != nil {
		return
	}

	keys = make([]string, 0)
	values = make([]*x509.Certificate, 0)
	keyToIndex = make(map[string]int)
	var cert *x509.Certificate
	for {
		var block *pem.Block
		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return
		}

		hash := sha256.Sum256(cert.Raw)
		fingerprint := hex.EncodeToString(hash[:])

		keys = append(keys, fingerprint)
		values = append(values, cert)
		keyToIndex[fingerprint] = len(keys) - 1

		if len(pemData) == 0 {
			break
		}
	}

	return
}

func (c *CACert) Restore() (err error, removedCerts []*x509.Certificate) {
	originalPath := c.originalPath()
	if _, err = os.Stat(originalPath); err != nil {
		return
	}
	backupPath := c.backupPath()
	if _, err = os.Stat(backupPath); err != nil {
		return
	}
	tmpPath := c.tmpPath()
	if _, err = os.Stat(tmpPath); err == nil {
		err = fmt.Errorf("temporary file %s already exists", tmpPath)
		return
	}
	err = os.Rename(originalPath, tmpPath)
	if err != nil {
		return
	}
	err = os.Rename(backupPath, originalPath)
	if err != nil {
		_ = os.Rename(tmpPath, originalPath)
		return
	}
	revert := func() {
		_ = os.Rename(originalPath, backupPath)
		_ = os.Rename(tmpPath, originalPath)
		return
	}
	backupHashes, backupHashToIndex, backupCerts, err := readCertsFromFile(tmpPath)
	if err != nil {
		revert()
		return
	}
	originalHashes, _, _, err := readCertsFromFile(originalPath)
	if err != nil {
		revert()
		return
	}
	if err = os.Remove(tmpPath); err != nil {
		revert()
		return
	}
	originalHashesSet := mapset.NewSet[string](originalHashes...)
	backupHashesSet := mapset.NewSet[string](backupHashes...)
	removedHashes := backupHashesSet.Difference(originalHashesSet)
	removedCerts = make([]*x509.Certificate, removedHashes.Cardinality())
	for i, hash := range removedHashes.ToSlice() {
		index, _ := backupHashToIndex[hash]
		removedCerts[i] = backupCerts[index]
	}
	return
}

func (c *CACert) Append(certs []*x509.Certificate) (err error) {
	originalPath := c.originalPath()
	if _, err = os.Stat(originalPath); err != nil {
		return
	}

	file, err := os.OpenFile(originalPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	for _, cert := range certs {
		if err = launcherCommonCert.WriteAsPem(cert.Raw, file); err != nil {
			return
		}
	}
	return
}
