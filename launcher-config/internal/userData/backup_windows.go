package userData

import (
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
	"golang.org/x/sys/windows"
)

func basePath(gameId string) string {
	if gameId == common.GameAoE4 {
		path, err := windows.KnownFolderPath(windows.FOLDERID_Documents, 0)
		if err != nil {
			return filepath.Join(os.Getenv("USERPROFILE"), "Documents")
		}
		return path
	}
	return os.Getenv("USERPROFILE")
}
