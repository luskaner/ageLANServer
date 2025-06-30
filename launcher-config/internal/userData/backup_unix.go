//go:build !windows

package userData

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-common/steam"
)

func basePath(gameTitle common.GameTitle) string {
	return steam.UserProfilePath(gameTitle)
}
