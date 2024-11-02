//go:build !windows

package steam

import (
	"path"
)

func UserProfilePath(gameId string) string {
	suffix := ConfigPath()
	if suffix == "" {
		return ""
	}
	return path.Join(suffix, "steamapps", "compatdata", AppId(gameId), "pfx", "drive_c", "users", "steamuser")
}
