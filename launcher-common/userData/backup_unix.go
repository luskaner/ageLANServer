//go:build !windows

package userData

import "github.com/luskaner/ageLANServer/common/game/steam"

func basePath(gameId string) string {
	return steam.UserProfilePath(gameId)
}
