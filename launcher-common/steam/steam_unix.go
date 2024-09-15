//go:build !windows

package steam

import (
	"path"
)

func UserProfilePath() string {
	suffix := ConfigPath()
	if suffix == "" {
		return ""
	}
	return path.Join(suffix, "steamapps", "compatdata", appID, "pfx", "drive_c", "users", "steamuser")
}
