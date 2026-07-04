package exec

import "slices"

// Source: https://github.com/i3/i3/blob/next/i3-sensible-terminal (removed legacy ones and unsupported)
// Sorted by being default and popularity
var terminalApplications = []string{
	// Meta-terminal
	xdgTerminalExecApp,
	// Actual terminals
	"gnome-terminal",
	"konsole",
	"xfce4-terminal",
	"mate-terminal",
	"qterminal",
	"lxterminal",
	"terminator",
	"tilix",
	"terminology",
	"termit",
}

func terminalArgs() []arg {
	apps := append(slices.Clone(terminalApplications), terminalApps...)
	apps = append(apps, x11TerminalApps...)
	return baseTerminalArgs(apps)
}
