//go:build !windows

package game

import (
	"os"

	"mvdan.cc/sh/v3/shell"
)

func FirstExistingDir(dirs []string, dirFn func(string) string) string {
	return FirstExistingFile(dirs, dirFn, func(stat os.FileInfo) bool {
		return stat.IsDir()
	})
}

func FirstExistingFile(files []string, fileFn func(string) string, checkFn func(os.FileInfo) bool) string {
	if fileFn == nil {
		fileFn = func(s string) string { return s }
	}
	var stat os.FileInfo
	for _, dir := range files {
		convertedDir, err := shell.Expand(fileFn(dir), nil)
		if err != nil {
			continue
		}
		if stat, err = os.Stat(convertedDir); err == nil && checkFn(stat) {
			return convertedDir
		}
	}
	return ""
}
