//go:build !windows

package exec

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func IsAdmin() bool {
	return os.Geteuid() == 0
}

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
		argsReplace := strings.Join(args[joinArgsIndex:], " ")
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

func startCmd(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:     true,
	}
	return cmd.Start()
}
