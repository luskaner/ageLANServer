package exec

func terminalArgs() []string {
	script := `
		on run argv
			tell application "Terminal"
				activate
				do script (item 1 of argv)
			end tell
		end run
	`
	return osascriptArgs(script)
}

func visualAdminArgs() []string {
	script :=
		`on run argv
			set curDir to do shell script "pwd"
			set AppleScript's text item delimiters to " "
			set cmd to argv as string	
			try
				do shell script "cd " & quoted form of curDir & " && " & cmd with administrator privileges
			on error m number n
				error m number n
			end try
		end run`
	return osascriptArgs(script)
}

func osascriptArgs(script string) []string {
	return []string{
		`osascript`,
		`-e`,
		script,
	}
}
