package game

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	commonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor/exec"
)

const appNamePrefix = "Microsoft."
const appPublisherId = "8wekyb3d8bbwe"

func appNameSuffix(id string) string {
	switch id {
	case common.GameAoE1:
		return "Darwin"
	case common.GameAoE2:
		return "MSPhoenix"
	case common.GameAoE3:
		return "MSGPBoston"
	default:
		return ""
	}
}

func appName(id string) string {
	return appNamePrefix + appNameSuffix(id)
}

func isInstalledOnXbox(id string) bool {
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

func (exec CustomExecutor) GameProcesses() (steamProcess bool, xboxProcess bool) {
	steamProcess = true
	xboxProcess = true
	return
}

func (exec XboxExecutor) Execute(_ []string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{
		File:        fmt.Sprintf(`shell:appsfolder\%s_%s!App`, appName(exec.gameId), appPublisherId),
		Shell:       true,
		SpecialFile: true,
	}.Exec()
	return
}

func (exec XboxExecutor) GameProcesses() (steamProcess bool, xboxProcess bool) {
	xboxProcess = true
	return
}

func startUri(uri string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{File: uri, Shell: true, SpecialFile: true}.Exec()
	return
}
