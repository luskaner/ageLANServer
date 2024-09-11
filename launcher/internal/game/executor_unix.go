//go:build !windows

package game

import (
	commonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
)

// MicrosoftStoreExecutor is not supported on non-Windows platforms
func isInstalledOnMicrosoftStore() bool {
	return false
}
func (exec MicrosoftStoreExecutor) Execute(_ []string) (result *commonExecutor.Result) {
	// Should not be called
	return
}
func (exec MicrosoftStoreExecutor) GameProcesses() (steamProcess bool, microsoftStoreProcess bool) {
	return
}

func (exec CustomExecutor) GameProcesses() (steamProcess bool, microsoftStoreProcess bool) {
	steamProcess = true
	return
}

func startUri(uri string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{File: openCommand(), Args: []string{uri}, Shell: true, SpecialFile: true}.Exec()
	return
}
