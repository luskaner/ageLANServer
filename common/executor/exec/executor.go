package exec

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/luskaner/ageLANServer/common/executor"
)

type Options struct {
	File           string
	SpecialFile    bool
	Shell          bool
	UseWorkingPath bool
	AsAdmin        bool
	Wait           bool
	ShowWindow     bool
	Pid            bool
	ExitCode       bool
	GUI            bool
	Args           []string
	Stdout         io.Writer
	Stderr         io.Writer
}

type Result struct {
	Err      error
	ExitCode int
	Pid      uint32
}

func (result *Result) Success() bool {
	return result != nil && result.Err == nil && (result.Pid != 0 || result.ExitCode == 0)
}

func (options Options) Exec() (result *Result) {
	result = &Result{}
	if options.GUI && !options.ShowWindow {
		result.Err = errors.New("gui apps need to set showWindow as true")
		return
	}
	if (options.GUI || options.ShowWindow) && (options.Stdout != nil || options.Stderr != nil) {
		result.Err = errors.New("gui/showWindow is not compatible with stdout/stderr")
		return
	}
	if options.File == "" {
		result.Err = errors.New("no file specified")
		return
	}
	options.AsAdmin = options.AsAdmin && !executor.IsAdmin()
	if !options.SpecialFile {
		options.File = getExecutablePath(options.File)
	}
	return options.exec()
}

func (options Options) standardExec() (result *Result) {
	result = &Result{}
	err, cmd := execCustomExecutable(
		options.File,
		options.GUI,
		options.Wait,
		options.ShowWindow,
		!options.UseWorkingPath,
		options.Stdout,
		options.Stderr,
		options.Args...,
	)
	if options.ExitCode && cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	if options.Pid && cmd.ProcessState == nil {
		result.Pid = uint32(cmd.Process.Pid)
	}
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			err = nil
		}
	}
	result.Err = err
	return
}

func (options Options) String() string {
	allArgs := []string{options.File}
	allArgs = append(allArgs, options.Args...)
	return strings.Join(allArgs, " ")
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

func makeCommand(executable string, executableWorkingPath bool, stdout io.Writer, stderr io.Writer, arg ...string) *exec.Cmd {
	cmd := exec.Command(executable, arg...)
	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}
	if executableWorkingPath {
		cmd.Dir = filepath.Dir(executable)
	}
	return cmd
}

func execCustomExecutable(executable string, gui bool, wait bool, show bool, executableWorkingPath bool, stdout io.Writer, stderr io.Writer, arg ...string) (error, *exec.Cmd) {
	cmd := makeCommand(executable, executableWorkingPath, stdout, stderr, arg...)
	var err error
	configureCmd(cmd, wait, show, gui)
	if wait {
		err = cmd.Run()
	} else {
		err = cmd.Start()
	}
	return err, cmd
}
