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
