//go:build !windows && !darwin

package exec

import (
	"github.com/luskaner/aoe2DELanServer/launcher-common"
	"golang.org/x/term"
	"os"
)

func terminalArgs() []string {
	switch {
	case launcher_common.Ubuntu():
		return []string{"x-terminal-emulator", "-e"}
	case launcher_common.SteamOS():
		return []string{"konsole", "-e"}
	}
	return nil
}

func adminArgs() []string {
	if term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) {
		return []string{"sudo", "-EH"}
	}
	return []string{"pkexec", "--keep-cwd"}
}
