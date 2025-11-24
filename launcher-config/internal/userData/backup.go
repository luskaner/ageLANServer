package userData

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/luskaner/ageLANServer/common/logger"
	commonUserData "github.com/luskaner/ageLANServer/launcher-common/userData"
)

type Data struct {
	Path string
}

func (d Data) isolatedPath() string {
	if ok, transformedPath := commonUserData.TransformPath(d.Path, commonUserData.TypeActive, commonUserData.TypeServer); ok {
		return transformedPath
	}
	return ""
}

func (d Data) originalPath() string {
	if ok, transformedPath := commonUserData.TransformPath(d.Path, commonUserData.TypeActive, commonUserData.TypeBackup); ok {
		return transformedPath
	}
	return ""
}

func (d Data) switchPaths(backupPath string, currentPath string) (ok bool) {
	commonLogger.Printf("\tSwitching %s <-> %s\n", currentPath, backupPath)
	if _, err := os.Stat(backupPath); err == nil {
		return
	}

	absolutePath := d.Path
	var mode os.FileMode

	if _, err := os.Stat(absolutePath); errors.Is(err, fs.ErrNotExist) {
		oldParent := absolutePath
		newParent := filepath.Dir(oldParent)
		var info os.FileInfo
		for {
			if runtime.GOOS == "linux" {
				if oldParent == newParent {
					return
				}
			} else if newParent == "." {
				return
			}
			info, err = os.Stat(newParent)
			if err == nil {
				mode = info.Mode()
				commonLogger.Printf("\t\tCreating all path hierarchy: %s\n", absolutePath)
				if err = os.MkdirAll(absolutePath, mode); err != nil {
					return
				}
				break
			} else if errors.Is(err, fs.ErrNotExist) {
				oldParent = newParent
				newParent = filepath.Dir(newParent)
			} else {
				return
			}
		}
	} else if err != nil {
		return
	}

	commonLogger.Printf("\t\tRenaming/Moving %s to %s\n", absolutePath, backupPath)
	if err := os.Rename(absolutePath, backupPath); err != nil {
		return
	}

	var revertMethods []func() bool
	defer func() {
		for i := len(revertMethods) - 1; i >= 0; i-- {
			if !revertMethods[i]() {
				break
			}
		}
	}()

	revertMethods = append(revertMethods, func() bool {
		commonLogger.Printf("\t\tRenaming/Moving %s to %s\n", backupPath, absolutePath)
		return os.Rename(backupPath, absolutePath) == nil
	})

	if _, err := os.Stat(currentPath); errors.Is(err, fs.ErrNotExist) {
		if mode == 0 {
			var absInfo os.FileInfo
			if absInfo, err = os.Stat(backupPath); err == nil {
				mode = absInfo.Mode()
			} else {
				return
			}
		}
		commonLogger.Printf("\t\tMaking directory %s\n", currentPath)
		if err = os.Mkdir(currentPath, mode); err != nil {
			return
		}
	} else if err != nil {
		return
	}

	commonLogger.Printf("\t\tRenaming/Moving %s to %s\n", currentPath, absolutePath)
	if err := os.Rename(currentPath, absolutePath); err != nil {
		return
	}

	revertMethods = nil
	return true
}

func (d Data) Backup() bool {
	return d.switchPaths(d.originalPath(), d.isolatedPath())
}

func (d Data) Restore() bool {
	return d.switchPaths(d.isolatedPath(), d.originalPath())
}
