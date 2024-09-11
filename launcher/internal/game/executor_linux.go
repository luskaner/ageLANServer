package game

import "github.com/luskaner/aoe2DELanServer/launcher-common"

func openCommand() string {
	file := "open"
	if launcher_common.SteamOS() {
		file = "xdg-open"
	}
	return file
}
