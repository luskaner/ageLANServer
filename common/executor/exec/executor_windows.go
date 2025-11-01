package exec

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows"
)

func fixArgs(arg ...string) []string {
	for i := range arg {
		arg[i] = fmt.Sprintf(`"%s"`, strings.ReplaceAll(arg[i], `"`, `\"`))
	}
	return arg
}

func (options Options) exec() (result *Result) {
	shell := options.Shell || options.AsAdmin
	if shell {
		result = &Result{}
		if options.Stdout != nil || options.Stderr != nil {
			result.Err = errors.New("shell or elevating as admin are not compatible with capturing stdout/stderr")
			return
		}
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

		err, pid, exitCode := shellExecuteEx(verb, !options.Wait, options.File, !options.UseWorkingPath, options.Pid, showWindowInt, options.Args...)
		result.Err = err
		if options.Pid {
			result.Pid = pid
		}
		if options.ExitCode {
			result.ExitCode = exitCode
		}
	} else {
		return options.standardExec()
	}
	return
}

func configureCmd(cmd *exec.Cmd, wait bool, show bool, gui bool) {
	if gui {
		if !wait {
			cmd.SysProcAttr = &windows.SysProcAttr{
				CreationFlags: windows.DETACHED_PROCESS,
			}
		}
	} else if show {
		cmd.SysProcAttr = &windows.SysProcAttr{
			NoInheritHandles: true,
			CreationFlags:    windows.CREATE_NEW_CONSOLE,
		}
	} else if !wait {
		cmd.SysProcAttr = &windows.SysProcAttr{
			CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NO_WINDOW,
		}
	} else {
		cmd.SysProcAttr = &windows.SysProcAttr{
			CreationFlags: windows.CREATE_NO_WINDOW,
		}
	}
}
