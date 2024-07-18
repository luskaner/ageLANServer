package userData

import (
	"os"
	"path/filepath"
)

type Data struct {
	Path string
}

const finalPath = `Games\Age of Empires 2 DE`

func (d *Data) isolatedPath() string {
	return d.absolutePath() + `.lan`
}

func (d *Data) originalPath() string {
	return d.absolutePath() + `.bak`
}

func (d *Data) absolutePath() string {
	return filepath.Join(Path(), d.Path)
}

func Path() string {
	return filepath.Join(os.Getenv("USERPROFILE"), finalPath)
}

func (d *Data) switchPaths(backupPath string, currentPath string) bool {
	absolutePath := d.absolutePath()
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

func (d *Data) Backup() bool {
	return d.switchPaths(d.originalPath(), d.isolatedPath())
}

func (d *Data) Restore() bool {
	return d.switchPaths(d.isolatedPath(), d.originalPath())
}
