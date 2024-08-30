package game

import (
	commonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"golang.org/x/sys/windows/registry"
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

func steamInstallationPath() (path string) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer func(key registry.Key) {
		_ = key.Close()
	}(key)
	var val string
	val, _, err = key.GetStringValue("SteamPath")
	if err != nil {
		return
	}
	return val
}

func startUri(uri string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{File: uri, Shell: true, SpecialFile: true}.Exec()
	return
}
