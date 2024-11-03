//go:build !windows

package userData

import "github.com/luskaner/ageLANServer/launcher-common/steam"

func basePath(gameId string) string {
	return steam.UserProfilePath(gameId)
}
