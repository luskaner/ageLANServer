package internal

import (
	"log"
	"os"
)

const finalPath = `Games\Age of Empires 2 DE`
const metadataOriginalFolder = "metadata"
const metadataBackupFolder = metadataOriginalFolder + ".bak"
const metadataTempFolder = metadataOriginalFolder + ".tmp"

func baseFolder() string {
	return os.Getenv("USERPROFILE") + `\` + finalPath
}

func switchOriginalAndBackup(basePath string, metadataPath string, metadataBackupPath string) {
	metadataTempPath := basePath + metadataTempFolder
	err := os.Rename(metadataPath, metadataTempPath)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Rename(metadataBackupPath, metadataPath)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Rename(metadataTempPath, metadataBackupPath)
	if err != nil {
		log.Fatal(err)
	}
}

func Backup() {
	basePath := baseFolder()
	metadataPath := basePath + metadataOriginalFolder
	metadataPathInfo, err := os.Stat(metadataPath)

	if err != nil {
		return
	}

	metadataBackupPath := basePath + metadataBackupFolder

	if _, err := os.Stat(metadataBackupPath); err != nil {
		err = os.Rename(metadataPath, metadataBackupPath)
		if err != nil {
			log.Fatal(err)
		}
		err = os.Mkdir(metadataPath, metadataPathInfo.Mode())
		if err != nil {
			log.Fatal(err)
		}
	} else {
		switchOriginalAndBackup(basePath, metadataPath, metadataBackupPath)
	}
}

func Restore() {
	basePath := baseFolder()
	metadataPath := basePath + metadataOriginalFolder
	_, err := os.Stat(metadataPath)

	if err != nil {
		return
	}

	metadataBackupPath := basePath + metadataBackupFolder

	if _, err := os.Stat(metadataBackupPath); err != nil {
		return
	}

	switchOriginalAndBackup(basePath, metadataPath, metadataBackupPath)
}
