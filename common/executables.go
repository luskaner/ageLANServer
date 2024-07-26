package common

import "runtime"

// Server

const Server = "server"
const ServerGenCert = "genCert"

// Launcher

const LauncherWatcher = "watcher"
const LauncherConfig = "config"
const LauncherConfigAdmin = "config-admin"
const LauncherConfigAdminAgent = "config-admin-agent"

func GetExeFileName(name string) string {
	filename := name
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}
	return filename
}
