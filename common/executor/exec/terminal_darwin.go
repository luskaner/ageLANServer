package exec

import (
	"fmt"
	"os"
)

const kittyFmt = `/Applications/%citty.app/Contents/MacOS/kitty`

var terminalPaths = map[string][]string{
	kittyApp: {
		fmt.Sprintf(kittyFmt, 'k'),
		fmt.Sprintf(kittyFmt, 'K'),
	},
	alacrittyApp: {`/Applications/Alacritty.app/Contents/MacOS/alacritty`},
	ghostyApp:    {`/Applications/Ghostty.app/Contents/MacOS/ghostty`},
	weztermApp:   {`/Applications/WezTerm.app/Contents/MacOS/wezterm`},
	rioApp:       {`/Applications/rio.app/Contents/MacOS/rio`},
}

func terminalArgs() []arg {
	for app, paths := range terminalPaths {
		for _, path := range paths {
			if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
				return appArgs(app, path)
			}
		}
	}
	if terminalAppArgs := baseTerminalArgs(terminalApps); len(terminalAppArgs) > 0 {
		return terminalAppArgs
	}
	return baseTerminalArgs(x11TerminalApps)
}
