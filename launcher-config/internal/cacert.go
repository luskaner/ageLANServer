package internal

import (
	"crypto/x509"
	"fmt"
	"io"
	"os"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common/logger"
	launcherCommonCert "github.com/luskaner/ageLANServer/launcher-common/cert"
)

type CACert struct {
	launcherCommonCert.CA
}

func NewCACert(gameId string, gamePath string) *CACert {
	return &CACert{launcherCommonCert.NewCA(gameId, gamePath)}
}

func (c *CACert) Backup() (err error) {
	originalPath := c.OriginalPath()
	if _, err = os.Stat(originalPath); err != nil {
		return
	}
	backupPath := c.BackupPath()
	if _, err = os.Stat(backupPath); err == nil {
		return
	}
	commonLogger.Printf("Opening %s\n", originalPath)
	originalFile, err := os.Open(originalPath)
	if err != nil {
		return
	}
	defer func(originalFile *os.File) {
		_ = originalFile.Close()
	}(originalFile)
	commonLogger.Printf("Creating %s\n", backupPath)
	backupFile, err := os.Create(backupPath)
	if err != nil {
		return
	}
	defer func(backupFile *os.File) {
		_ = backupFile.Close()
	}(backupFile)
	commonLogger.Printf("Copying data from %s to %s\n", backupPath, originalPath)
	_, err = io.Copy(backupFile, originalFile)
	if err != nil {
		_ = backupFile.Close()
		_ = os.Remove(backupPath)
		return
	}

	_ = backupFile.Sync()
	return
}

func (c *CACert) Restore() (err error, removedCerts []*x509.Certificate) {
	originalPath := c.OriginalPath()
	if _, err = os.Stat(originalPath); err != nil {
		return
	}
	backupPath := c.BackupPath()
	if _, err = os.Stat(backupPath); err != nil {
		return
	}
	tmpPath := c.TmpPath()
	if _, err = os.Stat(tmpPath); err == nil {
		err = fmt.Errorf("temporary file %s already exists", tmpPath)
		return
	}
	commonLogger.Printf("Renaming/Moving %s to %s\n", originalPath, tmpPath)
	err = os.Rename(originalPath, tmpPath)
	if err != nil {
		return
	}
	commonLogger.Printf("Renaming/Moving %s to %s\n", backupPath, originalPath)
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
	commonLogger.Printf("Reading %s certificates\n", tmpPath)
	backupHashes, backupHashToIndex, backupCerts, err := launcherCommonCert.ReadFromFile(tmpPath)
	if err != nil {
		revert()
		return
	}
	commonLogger.Printf("Reading %s certificates\n", originalPath)
	originalHashes, _, _, err := launcherCommonCert.ReadFromFile(originalPath)
	if err != nil {
		revert()
		return
	}
	commonLogger.Printf("Deleting %s\n", tmpPath)
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
	originalPath := c.OriginalPath()
	if _, err = os.Stat(originalPath); err != nil {
		return
	}
	commonLogger.Printf("Opening %s\n", originalPath)
	file, err := os.OpenFile(originalPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	commonLogger.Println("Writing certs data")
	for _, cert := range certs {
		if err = launcherCommonCert.WriteAsPem(cert.Raw, file); err != nil {
			return
		}
	}
	return
}
