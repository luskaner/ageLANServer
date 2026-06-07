package exec

func terminalArgs() []arg {
	script := `
		on run argv
			set AppleScript's text item delimiters to " "
			set cmd to argv as string
			tell application "Terminal"
				activate
				do script cmd
			end tell
		end run
	`
	return osascriptArgs(script)
}

func visualAdminArgs() []arg {
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

func osascriptArgs(script string) []arg {
	return []arg{
		newSafeArg(`osascript`),
		newSafeArg(`-e`),
		newUnsafeArg(script),
	}
}
