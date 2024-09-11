package exec

import (
	"golang.org/x/sys/windows"
	"path/filepath"
	"strings"
)

func IsAdmin() bool {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	return token.IsElevated()
}

func shellExecute(verb string, file string, executableWorkingPath bool, showWindow int32, arg ...string) error {
	verbPtr, _ := windows.UTF16PtrFromString(verb)
	exe, _ := windows.UTF16PtrFromString(file)
	args, _ := windows.UTF16PtrFromString(strings.Join(arg, " "))
	var workingDirPtr *uint16
	if executableWorkingPath {
		workingDirPtr, _ = windows.UTF16PtrFromString(filepath.Dir(file))
	} else {
		workingDirPtr = nil
	}

	err := windows.ShellExecute(0, verbPtr, exe, args, workingDirPtr, showWindow)
	return err
}

func (options Options) exec() (result *Result) {
	shell := options.Shell || options.ShowWindow || options.AsAdmin || !options.Wait
	if shell {
		result = &Result{}
		var showWindowInt int32

		if options.ShowWindow {
			showWindowInt = windows.SW_NORMAL
		} else {
			showWindowInt = windows.SW_HIDE
		}

		var verb string
		switch {
		case options.AsAdmin:
			verb = "runas"
		default:
			verb = "open"
		}

		if options.Wait || options.Pid || options.ExitCode {
			err, pid, exitCode := shellExecuteEx(verb, !options.Wait, options.File, !options.UseWorkingPath, options.Pid, showWindowInt, options.Args...)
			result.Err = err
			if options.Pid {
				result.Pid = pid
			}
			if options.ExitCode {
				result.ExitCode = exitCode
			}
		} else {
			result.Err = shellExecute(verb, options.File, !options.UseWorkingPath, showWindowInt, options.Args...)
		}
	} else {
		return options.standardExec()
	}
	return
}
