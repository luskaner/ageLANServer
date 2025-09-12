package userData

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/luskaner/ageLANServer/common"
)

type Data struct {
	Path string
}

const finalPathPrefix = "Games"

func finalPath(gameId string) string {
	var suffix string
	switch gameId {
	case common.GameAoE1:
		suffix = filepath.Join(`Age of Empires DE`, `Users`)
	case common.GameAoE2:
		suffix = `Age of Empires 2 DE`
	case common.GameAoE3:
		suffix = `Age of Empires 3 DE`
	case common.GameAoM:
		suffix = `Age of Mythology Retold`
	}
	return filepath.Join(finalPathPrefix, suffix)
}

func (d Data) isolatedPath(gameId string) string {
	return d.absolutePath(gameId) + `.lan`
}

func (d Data) originalPath(gameId string) string {
	return d.absolutePath(gameId) + `.bak`
}

func (d Data) absolutePath(gameId string) string {
	return filepath.Join(path(gameId), d.Path)
}

func path(gameId string) string {
	return filepath.Join(basePath(gameId), finalPath(gameId))
}

func (d Data) switchPaths(gameId, backupPath string, currentPath string) (ok bool) {
	if _, err := os.Stat(backupPath); err == nil {
		return
	}

	absolutePath := d.absolutePath(gameId)
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

func (d Data) Backup(gameId string) bool {
	return d.switchPaths(gameId, d.originalPath(gameId), d.isolatedPath(gameId))
}

func (d Data) Restore(gameId string) bool {
	return d.switchPaths(gameId, d.isolatedPath(gameId), d.originalPath(gameId))
}
