//go:build !windows

package steam

import (
	"os"
	"path"
)

func HomeDirPath() (p string) {
	if homeDir, err := os.UserHomeDir(); err == nil {
		p = path.Join(homeDir, dir)
	}
	if info, err := os.Stat(p); err != nil || !info.IsDir() {
		p = ""
	}
	return
}

func UserProfilePath() string {
	return path.Join(dir, "steamapps", "compatdata", appID, "pfx", "drive_c", "users", "steamuser")
}
