package common

import (
	"path/filepath"
)

// Server

const Server = "server"
const ServerGenCert = "genCert"

// Launcher

const LauncherAgent = "agent"
const LauncherConfig = "config"
const LauncherConfigAdmin = "config-admin"
const LauncherConfigAdminAgent = "config-admin-agent"

func GetExeFileName(bin bool, executable string) string {
	filename := getExeFileName(executable)
	if !bin {
		filename = filepath.Join("bin", filename)
	}
	return filename
}
