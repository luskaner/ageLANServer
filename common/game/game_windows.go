package game

import (
	"os"

	"github.com/luskaner/ageLANServer/common"
	"golang.org/x/sys/windows"
)

func UserProfilePath(gameId string) string {
	if gameId == common.GameAoE4 {
		if path, err := windows.KnownFolderPath(windows.FOLDERID_Documents, 0); err == nil {
			return path
		}
		return ""
	}
	return os.Getenv("USERPROFILE")
}
