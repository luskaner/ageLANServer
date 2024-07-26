package common

import (
	"path/filepath"
	"runtime"
)

// Server

const Server = "server"
const ServerGenCert = "genCert"

// Launcher

const LauncherWatcher = "watcher"
const LauncherConfig = "config"
const LauncherConfigAdmin = "config-admin"
const LauncherConfigAdminAgent = "config-admin-agent"

func getExeFileName(name string) string {
	filename := name
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}
	return filename
}

func GetExeFileName(bin bool, executable string) string {
	filename := getExeFileName(executable)
	if !bin {
		filename = filepath.Join("bin", filename)
	}
	return filename
}
