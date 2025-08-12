//go:build !windows

package game

import (
	"github.com/luskaner/ageLANServer/common"
	commonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor/exec"
)

// XboxExecutor is not supported on non-Windows platforms
func isInstalledOnXbox(_ common.GameTitle) bool {
	return false
}
func (exec XboxExecutor) Execute(_ []string) (result *commonExecutor.Result) {
	// Should not be called
	return
}
func (exec XboxExecutor) GameProcesses() (steamProcess bool, xboxProcess bool) {
	return
}

func (exec CustomExecutor) GameProcesses() (steamProcess bool, xboxProcess bool) {
	steamProcess = true
	return
}

func startUri(uri string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{File: openCommand(), Args: []string{uri}, SpecialFile: true, Pid: true}.Exec()
	return
}
