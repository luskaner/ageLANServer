//go:build !windows

package steam

import (
	"github.com/luskaner/ageLANServer/common"
	"path"
)

func UserProfilePath(gameTitle common.GameTitle) string {
	suffix := ConfigPath()
	if suffix == "" {
		return ""
	}
	return path.Join(suffix, "steamapps", "compatdata", AppId(gameTitle), "pfx", "drive_c", "users", "steamuser")
}
