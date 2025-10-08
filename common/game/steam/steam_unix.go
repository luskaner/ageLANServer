//go:build !windows

package steam

import (
	"path/filepath"
)

func UserProfilePath(gameId string) string {
	suffix := ConfigPath()
	if suffix == "" {
		return ""
	}
	return filepath.Join(suffix, "steamapps", "compatdata", AppId(gameId), "pfx", "drive_c", "users", "steamuser")
}
