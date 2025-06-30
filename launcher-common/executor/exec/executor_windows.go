package exec

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func fixArgs(arg ...string) []string {
	for i := range arg {
		arg[i] = fmt.Sprintf(`"%s"`, strings.ReplaceAll(arg[i], `"`, `\"`))
	}
	return arg
}

func shellExecute(verb string, file string, executableWorkingPath bool, showWindow int32, arg ...string) error {
	verbPtr, _ := windows.UTF16PtrFromString(verb)
	exe, _ := windows.UTF16PtrFromString(file)
	args, _ := windows.UTF16PtrFromString(strings.Join(fixArgs(arg...), " "))
	var workingDirPtr *uint16
	if executableWorkingPath {
		workingDirPtr, _ = windows.UTF16PtrFromString(filepath.Dir(file))
	} else {
		workingDirPtr = nil
	}

	return windows.ShellExecute(0, verbPtr, exe, args, workingDirPtr, showWindow)
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

		if len(options.Env) > 0 {
			current := environ()
			var revertMethods []func()
			var err error
			for k, v := range options.Env {
				existing, ok := current[k]
				if !ok || existing != v {
					if err = os.Setenv(k, v); err != nil {
						break
					}
					if ok {
						revertMethods = append(revertMethods, func() {
							_ = os.Setenv(k, existing)
						})
					} else {
						revertMethods = append(revertMethods, func() {
							_ = os.Unsetenv(k)
						})
					}
				}
			}
			defer func() {
				for _, revert := range revertMethods {
					revert()
				}
			}()
		}
		if options.Wait || options.Pid || options.ExitCode {
			err, pid, exitCode := shellExecuteEx(verb, !options.Wait, options.File, !options.UseWorkingPath, options.Pid, showWindowInt, options.Args...)
			result.Err = err
			if options.Pid {
				result.Pid = int(pid)
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

func startCmd(cmd *exec.Cmd) error {
	return cmd.Start()
}
