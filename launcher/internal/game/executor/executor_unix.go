//go:build !windows

package executor

import (
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
)

func (exec XboxExec) Do(_ []string, _ func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	// Should not be called
	return
}
func (exec XboxExec) GameProcesses() (steamProcess bool, xboxProcess bool) {
	return
}
func (exec XboxExec) GamePath() string { return "" }

func (exec CustomExec) GameProcesses() (steamProcess bool, xboxProcess bool) {
	steamProcess = true
	return
}

func startUri(uri string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	options := commonExecutor.Options{File: openCommand(), Args: []string{uri}, SpecialFile: true, Pid: true}
	optionsFn(options)
	result = options.Exec()
	return
}
