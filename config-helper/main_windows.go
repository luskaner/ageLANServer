package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/game/steam"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run executes the requested subcommand. Errors are returned instead of being
// silently swallowed: an empty stdout with a zero exit code would be
// indistinguishable from a successfully converted empty path.
func run(args []string) error {
	if len(args) < 2 {
		return errors.New("usage: config-helper <windowsToUnixPath|configPath|userProfilePath> [args...]")
	}
	extra := args[2:]
	switch args[1] {
	case "windowsToUnixPath":
		if len(extra) < 1 {
			return errors.New("windowsToUnixPath requires a path argument")
		}
		return convertAndPrint(extra[0])
	case "configPath":
		if len(extra) < 1 {
			return errors.New("configPath requires a boolean argument")
		}
		var result string
		if extra[0] == "true" {
			result = steam.ConfigPathAlt()
		} else {
			result = steam.ConfigPath()
		}
		return convertAndPrint(result)
	case "userProfilePath":
		if len(extra) < 1 {
			return errors.New("userProfilePath requires a profile argument")
		}
		return convertAndPrint(game.UserProfilePath(extra[0]))
	default:
		return fmt.Errorf("unknown command %q", args[1])
	}
}

func convertAndPrint(path string) error {
	convertedResult, err := WindowsToUnixPath(path)
	if err != nil {
		return err
	}
	fmt.Print(convertedResult)
	return nil
}
