//go:build !windows && !darwin

package exec

import (
	"github.com/hairyhenderson/go-which"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/shell"
	"os"
)

// Source: https://github.com/i3/i3/blob/next/i3-sensible-terminal
var terminalApplications = []string{
	"$TERMINAL",
	"x-terminal-emulator",
	"mate-terminal",
	"gnome-terminal",
	"terminator",
	"xfce4-terminal",
	"urxvt",
	"rxvt",
	"termit",
	"Eterm",
	"aterm",
	"uxterm",
	"xterm",
	"roxterm",
	"termite",
	"lxterminal",
	"terminology",
	"st",
	"qterminal",
	"lilyterm",
	"tilix",
	"terminix",
	"konsole",
	"kitty",
	"guake",
	"tilda",
	"alacritty",
	"hyper",
	"wezterm",
	"rio",
}

func terminalArgs() []string {
	var terminal string
	for _, executable := range terminalApplications {
		expandedTerminal, err := shell.Expand(executable, nil)
		if err != nil {
			continue
		}
		terminal = which.Which(expandedTerminal)
		if terminal != "" {
			break
		}
	}
	if terminal == "" {
		return []string{}
	}
	return []string{terminal, "-e"}
}

func adminArgs(wait bool) []string {
	if !wait || !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return []string{"pkexec", "--keep-cwd"}
	}
	return []string{"sudo", "-EH"}
}
