//go:build !windows

package game

import (
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
)

func (exec XboxExecutor) Execute(_ []string) (result *commonExecutor.Result) {
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

func startUri(uri string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{File: openCommand(), Args: []string{uri}, SpecialFile: true, Pid: true}.Exec()
	return
}
