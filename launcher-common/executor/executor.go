package executor

import (
	"common"
	"errors"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/charmap"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ExecOptions struct {
	File           string
	SpecialFile    bool
	UseWorkingPath bool
	AsAdmin        bool
	Wait           bool
	WindowState    int
	Pid            bool
	ExitCode       bool
	Output         bool
	Args           []string
}

type ExecResult struct {
	Err      error
	Output   *string
	ExitCode int
	Pid      uint32
}

func (result *ExecResult) Success() bool {
	return result != nil && result.Err == nil && (result.Pid != 0 || result.ExitCode == common.ErrSuccess)
}

func (options ExecOptions) Exec() (result *ExecResult) {
	result = &ExecResult{}
	if options.File == "" {
		result.Err = errors.New("no file specified")
		return
	}
	options.AsAdmin = options.AsAdmin && !IsAdmin()
	shell := options.WindowState != windows.SW_HIDE || options.SpecialFile || options.AsAdmin
	if (shell || !options.Wait) && options.Output {
		result.Err = errors.New("output is not supported for shell or non-waiting processes")
		return
	}
	fullFile := options.File
	if shell {
		var verb string
		switch {
		case options.AsAdmin:
			verb = "runas"
		default:
			verb = "open"
		}
		if !options.SpecialFile {
			fullFile = getExecutablePath(fullFile)
		}
		if options.Wait || options.Pid || options.ExitCode {
			err, pid, exitCode := shellExecuteEx(verb, !options.Wait, fullFile, !options.UseWorkingPath, options.WindowState, options.Args...)
			result.Err = err
			if options.Pid {
				result.Pid = pid
			}
			if options.ExitCode {
				result.ExitCode = exitCode
			}
		} else {
			result.Err = shellExecute(verb, fullFile, !options.UseWorkingPath, options.WindowState, options.Args...)
		}
	} else {
		fullFile = getExecutablePath(fullFile)
		if options.Output {
			result.Err, result.Output = runCustomExecutableOutput(fullFile, !options.UseWorkingPath, options.Args...)
		} else {
			err, cmd := execCustomExecutable(fullFile, !options.Wait, !options.UseWorkingPath, options.Args...)
			if options.Pid {
				result.Pid = uint32(cmd.Process.Pid)
			}
			if options.ExitCode && cmd.ProcessState != nil {
				result.ExitCode = cmd.ProcessState.ExitCode()
			}
			if err != nil {
				var exitError *exec.ExitError
				if errors.As(err, &exitError) {
					err = nil
				}
			}
			result.Err = err
		}
	}
	return
}

func getExecutablePath(executable string) string {
	if filepath.IsLocal(executable) {
		ex, err := os.Executable()
		if err != nil {
			return ""
		}
		return filepath.Join(filepath.Dir(ex), executable)
	}
	return executable
}

func makeCommand(executable string, executableWorkingPath bool, arg ...string) *exec.Cmd {
	cmd := exec.Command(executable, arg...)
	if executableWorkingPath {
		cmd.Dir = filepath.Dir(executable)
	}
	return cmd
}

func execCustomExecutable(executable string, start bool, executableWorkingPath bool, arg ...string) (error, *exec.Cmd) {
	cmd := makeCommand(executable, executableWorkingPath, arg...)
	var err error
	if start {
		err = cmd.Start()
	} else {
		err = cmd.Run()
	}
	return err, cmd
}

func runCustomExecutableOutput(executable string, executableWorkingPath bool, arg ...string) (err error, output *string) {
	cmd := makeCommand(executable, executableWorkingPath, arg...)
	var outputRaw []byte
	outputRaw, err = cmd.Output()
	if err != nil {
		return
	}
	decoder := charmap.CodePage437.NewDecoder()
	out, err := decoder.Bytes(outputRaw)
	if err != nil {
		return
	}
	outStr := string(out)
	output = &outStr
	return
}

func IsAdmin() bool {
	err, token := GetCurrentProcessToken()
	if err != nil {
		return false
	}
	return token.IsElevated()
}

func GetCurrentProcessToken() (err error, token windows.Token) {
	err = windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	return
}

func shellExecute(verb string, file string, executableWorkingPath bool, showWindow int, arg ...string) error {
	verbPtr, _ := windows.UTF16PtrFromString(verb)
	exe, _ := windows.UTF16PtrFromString(file)
	args, _ := windows.UTF16PtrFromString(strings.Join(arg, " "))
	var workingDirPtr *uint16
	if executableWorkingPath {
		workingDirPtr, _ = windows.UTF16PtrFromString(filepath.Dir(file))
	} else {
		workingDirPtr = nil
	}

	show := showWindow

	err := windows.ShellExecute(0, verbPtr, exe, args, workingDirPtr, int32(show))
	return err
}
