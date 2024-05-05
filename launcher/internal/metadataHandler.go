package internal

import (
	"os"
)

const finalPath = `Games\Age of Empires 2 DE`
const metadataOriginalFolder = "metadata"
const metadataBackupFolder = metadataOriginalFolder + ".bak"
const metadataTempFolder = metadataOriginalFolder + ".tmp"

func baseFolder() string {
	return os.Getenv("USERPROFILE") + `\` + finalPath + `\`
}

func switchOriginalAndBackup(basePath string, metadataPath string, metadataBackupPath string) bool {
	metadataTempPath := basePath + metadataTempFolder
	err := os.Rename(metadataPath, metadataTempPath)
	if err != nil {
		return false
	}
	err = os.Rename(metadataBackupPath, metadataPath)
	if err != nil {
		return false
	}
	err = os.Rename(metadataTempPath, metadataBackupPath)
	if err != nil {
		return false
	}
	return true
}

func BackupMetadata() bool {
	basePath := baseFolder()
	metadataPath := basePath + metadataOriginalFolder
	metadataPathInfo, err := os.Stat(metadataPath)

	if err != nil {
		return false
	}

	metadataBackupPath := basePath + metadataBackupFolder

	if _, err := os.Stat(metadataBackupPath); err != nil {
		err = os.Rename(metadataPath, metadataBackupPath)
		if err != nil {
			return false
		}
		err = os.Mkdir(metadataPath, metadataPathInfo.Mode())
		if err != nil {
			return false
		}
		return true
	} else {
		return switchOriginalAndBackup(basePath, metadataPath, metadataBackupPath)
	}
}

func RestoreMetadata() bool {
	basePath := baseFolder()
	metadataPath := basePath + metadataOriginalFolder
	_, err := os.Stat(metadataPath)

	if err != nil {
		return false
	}

	metadataBackupPath := basePath + metadataBackupFolder

	if _, err := os.Stat(metadataBackupPath); err != nil {
		return false
	}

	return switchOriginalAndBackup(basePath, metadataPath, metadataBackupPath)
}
