package parentCheck

import (
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
)

func ParentMatches() bool {
	path := common.GetExeFileName(true, common.LauncherConfig)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	ppid := os.Getppid()
	exePath, err := exePathFromPID(ppid)
	if err != nil {
		return false
	}

	return exePath == absPath
}
