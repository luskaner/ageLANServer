//go:build !windows

package exec

import (
	"github.com/hairyhenderson/go-which"
)

var x11TerminalApps = []string{
	"xterm",
	"uxterm",
	"urxvt",
	"st",
}

const xdgTerminalExecApp = "xdg-terminal-exec"
const kittyApp = "kitty"
const alacrittyApp = "alacritty"
const ghostyApp = "ghostty"
const weztermApp = "wezterm"
const rioApp = "rio"

var terminalApps = []string{
	kittyApp,
	alacrittyApp,
	ghostyApp,
	weztermApp,
	rioApp,
}

var defExecArgs = []string{"-e"}

var execArgsPerApp = map[string][]string{
	xdgTerminalExecApp: {},
	kittyApp:           {},
}

func appArgs(app string, path string) []arg {
	var execArgs []string
	var ok bool
	if execArgs, ok = execArgsPerApp[app]; !ok {
		execArgs = defExecArgs
	}
	args := make([]arg, len(execArgs)+1)
	args[0] = newSafeArg(path)
	for i, argument := range execArgs {
		args[i+1] = newSafeArg(argument)
	}
	return args
}

func baseTerminalArgs(apps []string) []arg {
	for _, app := range apps {
		path := which.Which(app)
		if path != "" {
			return appArgs(app, path)
		}
	}
	return []arg{}
}
