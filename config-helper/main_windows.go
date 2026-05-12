package main

import (
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/game/steam"
)

func main() {
	switch os.Args[1] {
	case "windowsToUnixPath":
		windowsToUnixPath(os.Args[2:]...)
	case "configPath":
		configPath(os.Args[2:]...)
	case "userProfilePath":
		userProfilePath(os.Args[2:]...)
	}
}

func windowsToUnixPath(args ...string) {
	if convertedResult, err := WindowsToUnixPath(args[0]); err == nil {
		fmt.Print(convertedResult)
	}
}

func configPath(args ...string) {
	var result string
	if args[0] == "true" {
		result = steam.ConfigPathAlt()
	} else {
		result = steam.ConfigPath()
	}
	windowsToUnixPath(result)
}

func userProfilePath(args ...string) {
	windowsToUnixPath(game.UserProfilePath(args[0]))
}
