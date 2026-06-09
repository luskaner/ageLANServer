package exec

func terminalArgs() []arg {
	script := `
		on run argv
			set AppleScript's text item delimiters to " "
			set cmd to argv as string	
			set tmpFile to do shell script "mktemp /tmp/ageLANServer_osascript_exit_code.XXXXXX"
			set fullCmd to cmd & "; echo $? > " & tmpFile
		
			tell application "Terminal"
				set w to (do script fullCmd)
				set winID to id of w
			end tell
		
			tell application "System Events" to tell process "Terminal"
				set miniaturized of (first window whose id is winID) to true
			end tell
		
			tell application "Terminal"
				repeat while busy of w is true
					delay 0.1
				end repeat
				close w
			end tell
		
			set exitCode to do shell script "cat " & tmpFile & "; rm " & tmpFile
			return exitCode
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
