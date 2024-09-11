//go:build !windows

package exec

import "os"

func IsAdmin() bool {
	return os.Geteuid() == 0
}

func (options Options) exec() (result *Result) {
	var args []string
	if options.ShowWindow {
		args = append(args, terminalArgs()...)
	}
	if options.AsAdmin {
		args = append(args, adminArgs()...)
	}
	if shell := options.Shell || options.SpecialFile; shell {
		args = append(args, shellArgs()...)
	}
	options.Args = append(args, options.Args...)
	return options.standardExec()
}

func shellArgs() []string {
	return []string{"sh", "-c"}
}
