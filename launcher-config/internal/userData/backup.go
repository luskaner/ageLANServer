package userData

import (
	"github.com/luskaner/aoe2DELanServer/common"
	"os"
	"path/filepath"
)

type Data struct {
	Path string
}

const finalPathPrefix = "Games"

func finalPath(gameId string) string {
	var suffix string
	switch gameId {
	case common.GameAoE2:
		suffix = `Age of Empires 2 DE`
	case common.GameAoE3:
		suffix = `Age of Empires 3 DE`
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
	return filepath.Join(Path(gameId), d.Path)
}

func Path(gameId string) string {
	return filepath.Join(basePath(gameId), finalPath(gameId))
}

func (d Data) switchPaths(gameId, backupPath string, currentPath string) bool {
	absolutePath := d.absolutePath(gameId)
	var mode os.FileMode

	if _, err := os.Stat(absolutePath); err != nil {
		parent := absolutePath
		for {
			parent = filepath.Dir(parent)
			if _, err = os.Stat(parent); !os.IsNotExist(err) {
				break
			}
		}
		var info os.FileInfo
		if info, err = os.Stat(parent); err == nil {
			mode = info.Mode()
			if err = os.MkdirAll(absolutePath, mode); err != nil {
				return false
			}
		} else {
			return false
		}
	}

	var revertMethods []func() bool
	defer func() {
		for i := len(revertMethods) - 1; i >= 0; i-- {
			if !revertMethods[i]() {
				break
			}
		}
	}()

	if _, err := os.Stat(backupPath); err == nil {
		return false
	}

	if err := os.Rename(absolutePath, backupPath); err != nil {
		return false
	} else {
		revertMethods = append(revertMethods, func() bool {
			return os.Rename(backupPath, absolutePath) == nil
		})
	}

	if _, err := os.Stat(currentPath); err != nil {
		if mode == 0 {
			var absInfo os.FileInfo
			if absInfo, err = os.Stat(backupPath); err == nil {
				mode = absInfo.Mode()
			} else {
				return false
			}
		}
		if err = os.Mkdir(currentPath, mode); err != nil {
			return false
		}
	}

	if err := os.Rename(currentPath, absolutePath); err != nil {
		revertMethods = append(revertMethods, func() bool {
			return os.Rename(absolutePath, currentPath) == nil
		})
		return false
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
