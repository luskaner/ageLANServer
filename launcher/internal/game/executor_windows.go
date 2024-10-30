package game

import (
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	commonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
)

const appNamePrefix = "Microsoft."
const appPublisherId = "8wekyb3d8bbwe"

func appNameSuffix(id string) string {
	switch id {
	case common.GameAoE2:
		return "MSPhoenix"
	case common.GameAoE3:
		// TODO: Check if this is correct
		return "MSBoston"
	default:
		return ""
	}
}

func appName(id string) string {
	return appNamePrefix + appNameSuffix(id)
}

func isInstalledOnMicrosoftStore(id string) bool {
	// Does not seem there is another way without cgo?
	return commonExecutor.Options{
		File:        "powershell",
		SpecialFile: true,
		Wait:        true,
		ExitCode:    true,
		Args: []string{
			"-Command",
			fmt.Sprintf("if ((Get-AppxPackage).Name -eq '%s') { exit 0 } else { exit 1 }", appName(id)),
		},
	}.Exec().Success()
}

func (exec CustomExecutor) GameProcesses() (steamProcess bool, microsoftStoreProcess bool) {
	steamProcess = true
	microsoftStoreProcess = true
	return
}

func (exec MicrosoftStoreExecutor) Execute(_ []string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{
		File:        fmt.Sprintf(`shell:appsfolder\%s_%s!App`, appName(exec.gameId), appPublisherId),
		Shell:       true,
		SpecialFile: true,
	}.Exec()
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
