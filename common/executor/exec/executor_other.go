//go:build !windows

package exec

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"mvdan.cc/sh/v3/syntax"
)

func (options Options) exec() (result *Result) {
	var args []string
	joinArgsIndex := -1
	if options.ShowWindow {
		args = append(args, terminalArgs()...)
	}
	if options.AsAdmin {
		args = append(args, adminArgs(options.Wait)...)
	}
	if shell := options.Shell || options.ShowWindow; shell {
		args = append(args, shellArgs()...)
		joinArgsIndex = len(args)
		if !options.UseWorkingPath {
			args = append(args, []string{"cd", filepath.Dir(options.File), "&&"}...)
		}
	}
	args = append(args, options.File)
	args = append(args, options.Args...)
	if joinArgsIndex != -1 {
		argsQuoted := make([]string, len(args)-joinArgsIndex)
		for i, arg := range args[joinArgsIndex:] {
			if quoted, err := syntax.Quote(arg, syntax.LangPOSIX); err == nil {
				argsQuoted[i] = quoted
			} else {
				return &Result{
					Err: fmt.Errorf("error quoting argument: %w", err),
				}
			}
		}
		argsReplace := strings.Join(argsQuoted, " ")
		args = args[:joinArgsIndex]
		args = append(args, argsReplace)
	}
	options.File = args[0]
	if len(args) > 1 {
		options.Args = args[1:]
	}
	return options.standardExec()
}

func shellArgs() []string {
	return []string{"sh", "-c"}
}

func configureCmd(cmd *exec.Cmd, wait bool, _ bool, _ bool) {
	if wait {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}
	}
}

func adminArgs(wait bool) []string {
	if !wait || !common.Interactive() {
		return visualAdminArgs()
	}
	return []string{"sudo", "-EH"}
}
