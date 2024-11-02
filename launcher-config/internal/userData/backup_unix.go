//go:build !windows

package userData

import "github.com/luskaner/aoe2DELanServer/launcher-common/steam"

func basePath(gameId string) string {
	return steam.UserProfilePath(gameId)
}
