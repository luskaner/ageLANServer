package userData

import (
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
	"golang.org/x/sys/windows"
)

func basePath(gameId string) string {
	if gameId == common.GameAoE4 {
		if path, err := windows.KnownFolderPath(windows.FOLDERID_Documents, 0); err == nil {
			return filepath.Join(path, "My Games")
		}
		return ""
	}
	return os.Getenv("USERPROFILE")
}
