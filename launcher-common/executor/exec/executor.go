package exec

import (
	"errors"
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	Env            map[string]string
}

type Result struct {
	Err      error
	ExitCode int
	Pid      int
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
	err, cmd := execCustomExecutable(options.File, options.Wait, !options.UseWorkingPath, options.Env, options.Args...)
	if options.ExitCode && cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	if options.Pid && cmd.ProcessState == nil {
		result.Pid = cmd.Process.Pid
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

func environ() map[string]string {
	env := os.Environ()
	envMap := make(map[string]string, len(env))
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		} else if len(parts) == 1 {
			envMap[parts[0]] = ""
		}
	}
	return envMap
}

func mapEnvToSlice(env map[string]string) []string {
	envSlice := make([]string, 0, len(env))
	for key, value := range env {
		if value == "" {
			envSlice = append(envSlice, key)
		} else {
			envSlice = append(envSlice, fmt.Sprintf("%s=%s", key, value))
		}
	}
	return envSlice
}

func makeCommand(executable string, executableWorkingPath bool, env map[string]string, arg ...string) *exec.Cmd {
	cmd := exec.Command(executable, arg...)
	if len(env) > 0 {
		current := environ()
		maps.Copy(current, env)
		cmd.Env = mapEnvToSlice(current)
	}
	if executableWorkingPath {
		cmd.Dir = filepath.Dir(executable)
	}
	return cmd
}

func execCustomExecutable(executable string, wait bool, executableWorkingPath bool, env map[string]string, arg ...string) (error, *exec.Cmd) {
	cmd := makeCommand(executable, executableWorkingPath, env, arg...)
	var err error
	if wait {
		err = cmd.Run()
	} else {
		err = startCmd(cmd)
	}
	return err, cmd
}
