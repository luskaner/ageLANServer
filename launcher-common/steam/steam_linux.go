package steam

import (
	"mvdan.cc/sh/v3/shell"
	"os"
	"strings"
)

const suffixDir = ".steam/steam"

var dirs = []string{
	// Official
	"$HOME/{suffixDir}",
	// Alternative official
	"$HOME/.local/share/Steam",
	// Snap
	"$HOME/snap/steam/common/{suffixDir}",
	// Flatpak
	"$HOME/.var/app/com.valvesoftware.Steam/{suffixDir}",
	// Less common from https://github.com/lutris/lutris/blob/master/lutris/util/steam/config.py
	"$HOME/.steam/debian-installation",
	"$HOME/.steam",
	"$HOME/.local/share/steam",
	"$HOME/snap/steam/common/.local/share/Steam",
	"$HOME/.var/app/com.valvesoftware.Steam/data/Steam",
	"/usr/share/steam",
	"/usr/local/share/steam",
}

func ConfigPath() string {
	var stat os.FileInfo
	for _, dir := range dirs {
		convertedDir, err := shell.Expand(strings.ReplaceAll(dir, "{suffixDir}", suffixDir), nil)
		if err != nil {
			continue
		}
		if stat, err = os.Stat(convertedDir); err == nil && stat.IsDir() {
			return convertedDir
		}
	}
	return ""
}
