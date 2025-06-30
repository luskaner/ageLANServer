package userData

import (
	"errors"
	"github.com/luskaner/ageLANServer/common"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

type Data struct {
	Path string
}

const finalPathPrefix = "Games"

func finalPath(gameTitle common.GameTitle) string {
	var suffix string
	switch gameTitle {
	case common.AoE1:
		suffix = filepath.Join(`Age of Empires DE`, `Users`)
	case common.AoE2:
		suffix = `Age of Empires 2 DE`
	case common.AoE3:
		suffix = `Age of Empires 3 DE`
	}
	return filepath.Join(finalPathPrefix, suffix)
}

func (d Data) isolatedPath(staticBasePath string, gameTitle common.GameTitle) string {
	return d.absolutePath(staticBasePath, gameTitle) + `.lan`
}

func (d Data) originalPath(staticBasePath string, gameTitle common.GameTitle) string {
	return d.absolutePath(staticBasePath, gameTitle) + `.bak`
}

func (d Data) absolutePath(staticBasePath string, gameTitle common.GameTitle) string {
	return filepath.Join(path(staticBasePath, gameTitle), d.Path)
}

func path(staticBasePath string, gameTitle common.GameTitle) string {
	if staticBasePath == "" {
		staticBasePath = basePath(gameTitle)
	}
	return filepath.Join(staticBasePath, finalPath(gameTitle))
}

func (d Data) switchPaths(staticBasePath string, gameTitle common.GameTitle, backupPath string, currentPath string) (ok bool) {
	if _, err := os.Stat(backupPath); err == nil {
		return
	}

	absolutePath := d.absolutePath(staticBasePath, gameTitle)
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
		if err = os.Mkdir(currentPath, mode); err != nil {
			return
		}
	} else if err != nil {
		return
	}

	if err := os.Rename(currentPath, absolutePath); err != nil {
		return
	}

	revertMethods = nil
	return true
}

func (d Data) Backup(staticBasePath string, gameTitle common.GameTitle) bool {
	return d.switchPaths(staticBasePath, gameTitle, d.originalPath(staticBasePath, gameTitle), d.isolatedPath(staticBasePath, gameTitle))
}

func (d Data) Restore(staticBasePath string, gameTitle common.GameTitle) bool {
	return d.switchPaths(staticBasePath, gameTitle, d.isolatedPath(staticBasePath, gameTitle), d.originalPath(staticBasePath, gameTitle))
}
