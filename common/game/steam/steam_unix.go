//go:build !windows

package steam

import (
	"path/filepath"

	"github.com/luskaner/ageLANServer/common/game/wine"
)

func UserProfilePath(gameId string) string {
	suffix := ConfigPath()
	if suffix == "" {
		return ""
	}
	return filepath.Join(suffix, "steamapps", "compatdata", appId(gameId), "pfx", wine.UserProfile("steamuser"))
}
