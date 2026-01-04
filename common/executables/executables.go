package executables

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Server

const Server = "server"
const ServerGenCert = "genCert"

// Launcher

const Launcher = "launcher"
const LauncherAgent = "agent"
const LauncherConfig = "config"
const LauncherConfigAdmin = "config-admin"
const LauncherConfigAdminAgent = "config-admin-agent"

// Battle Server Manager

const BattleServerManager = "battle-server-manager"

var directories = []string{
	fmt.Sprintf(`%c`, filepath.Separator),
	fmt.Sprintf(`%c..%c`, filepath.Separator, filepath.Separator),
	fmt.Sprintf(`%c..%c..%c`, filepath.Separator, filepath.Separator, filepath.Separator),
}

func Filename(bin bool, executable string) string {
	filename := fileName(executable)
	if !bin {
		filename = filepath.Join("bin", filename)
	}
	return filename
}

func BaseNameNoExt(fileName string) string {
	extension := filepath.Ext(fileName)
	return strings.TrimSuffix(fileName, extension)
}

func FindPath(executableName string) string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	exeDir := BaseNameNoExt(executableName)
	exePath := filepath.Dir(ex)
	var f os.FileInfo
	for _, dir := range directories {
		executablePath := filepath.Join(exePath, dir, exeDir, executableName)
		if f, err = os.Stat(executablePath); err == nil && !f.IsDir() {
			executablePath, _ = filepath.Abs(executablePath)
			return executablePath
		}
	}
	return ""
}
