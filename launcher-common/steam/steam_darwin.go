package steam

import (
	"mvdan.cc/sh/v3/shell"
	"os"
)

const dir = "$HOME/Library/Application Support/Steam"

func ConfigPath() string {
	convertedDir, err := shell.Expand(dir, nil)
	if err != nil {
		return ""
	}
	var stat os.FileInfo
	if stat, err = os.Stat(convertedDir); err == nil && stat.IsDir() {
		return convertedDir
	}
	return ""
}
