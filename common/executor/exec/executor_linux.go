package exec

import "slices"

// Source: https://github.com/i3/i3/blob/next/i3-sensible-terminal (removed legacy ones)
var terminalApplications = []string{
	"$TERMINAL",
	"x-terminal-emulator",
	"mate-terminal",
	"gnome-terminal",
	"terminator",
	"xfce4-terminal",
	"termit",
	"lxterminal",
	"terminology",
	"qterminal",
	"tilix",
	"konsole",
	"guake",
	"tilda",
}

func visualAdminArgs() []arg {
	return []arg{newSafeArg("pkexec"), newSafeArg("--keep-cwd")}
}

func terminalArgs() []arg {
	apps := slices.Clone(terminalApps)
	apps = slices.Clone(terminalApplications)
	apps = append(apps, x11TerminalApps...)
	return baseTerminalArgs(apps)
}
