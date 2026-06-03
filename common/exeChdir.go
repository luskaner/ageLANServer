package common

import (
	"os"
	"path/filepath"
)

func ChdirToExe() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	exeDir := filepath.Dir(exePath)
	_ = os.Chdir(exeDir)
}
