package exec

import (
	"errors"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/common/executor"
	"os"
	"os/exec"
	"path/filepath"
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
	Args           []string
}

type Result struct {
	Err      error
	ExitCode int
	Pid      uint32
}

func (result *Result) Success() bool {
	return result != nil && result.Err == nil && (result.Pid != 0 || result.ExitCode == common.ErrSuccess)
}

func (options Options) Exec() (result *Result) {
	result = &Result{}
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
	err, cmd := execCustomExecutable(options.File, options.Wait, !options.UseWorkingPath, options.Args...)
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

func execCustomExecutable(executable string, wait bool, executableWorkingPath bool, arg ...string) (error, *exec.Cmd) {
	cmd := makeCommand(executable, executableWorkingPath, arg...)
	var err error
	if wait {
		err = cmd.Run()
	} else {
		err = startCmd(cmd)
	}
	return err, cmd
}
