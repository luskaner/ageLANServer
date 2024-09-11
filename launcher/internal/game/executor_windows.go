package game

import (
	commonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
)

func isInstalledOnMicrosoftStore() bool {
	// Does not seem there is another way without cgo?
	return commonExecutor.Options{File: "powershell", SpecialFile: true, Wait: true, ExitCode: true, Args: []string{"-Command", "if ((Get-AppxPackage).Name -eq 'Microsoft.MSPhoenix') { exit 0 } else { exit 1 }"}}.Exec().Success()
}

func (exec CustomExecutor) GameProcesses() (steamProcess bool, microsoftStoreProcess bool) {
	steamProcess = true
	microsoftStoreProcess = true
	return
}

func (exec MicrosoftStoreExecutor) Execute(_ []string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{File: `shell:appsfolder\Microsoft.MSPhoenix_8wekyb3d8bbwe!App`, Shell: true, SpecialFile: true}.Exec()
	return
}

func (exec MicrosoftStoreExecutor) GameProcesses() (steamProcess bool, microsoftStoreProcess bool) {
	microsoftStoreProcess = true
	return
}

func startUri(uri string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{File: uri, Shell: true, SpecialFile: true}.Exec()
	return
}
