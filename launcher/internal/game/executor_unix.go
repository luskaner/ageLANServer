//go:build !windows

package game

import (
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
)

func (exec XboxExecutor) Execute(_ []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	// Should not be called
	return
}
func (exec XboxExecutor) GameProcesses() (steamProcess bool, xboxProcess bool) {
	return
}
func (exec XboxExecutor) GamePath() string { return "" }

func (exec CustomExecutor) GameProcesses() (steamProcess bool, xboxProcess bool) {
	steamProcess = true
	return
}

func startUri(uri string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	options := commonExecutor.Options{File: openCommand(), Args: []string{uri}, SpecialFile: true, Pid: true}
	optionsFn(options)
	result = options.Exec()
	return
}
