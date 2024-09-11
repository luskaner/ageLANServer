package game

import (
	commonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"github.com/luskaner/aoe2DELanServer/launcher-common/steam"
	"os"
)

type Executor interface {
	Execute(args []string) (result *commonExecutor.Result)
	GameProcesses() (steamProcess bool, microsoftStoreProcess bool)
}

type SteamExecutor struct{}
type MicrosoftStoreExecutor struct{}
type CustomExecutor struct {
	Executable string
}

func (exec SteamExecutor) Execute(_ []string) (result *commonExecutor.Result) {
	return startUri(steam.OpenUri())
}

func (exec SteamExecutor) GameProcesses() (steamProcess bool, microsoftStoreProcess bool) {
	steamProcess = true
	return
}

func (exec CustomExecutor) Execute(args []string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{File: exec.Executable, Args: args}.Exec()
	return
}

func (exec CustomExecutor) ExecuteElevated(args []string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{File: exec.Executable, AsAdmin: true, ShowWindow: true, Args: args}.Exec()
	return
}

func isInstalledCustom(executable string) bool {
	info, err := os.Stat(executable)
	if err != nil || os.IsNotExist(err) || info.IsDir() {
		return false
	}
	return true
}

func MakeExecutor(executable string) Executor {
	if executable != "auto" {
		switch executable {
		case "steam":
			if steam.GameInstalled() {
				return SteamExecutor{}
			}
		case "msstore":
			if isInstalledOnMicrosoftStore() {
				return MicrosoftStoreExecutor{}
			}
		default:
			if isInstalledCustom(executable) {
				return CustomExecutor{Executable: executable}
			}
		}
		return nil
	}
	if steam.GameInstalled() {
		return SteamExecutor{}
	}
	if isInstalledOnMicrosoftStore() {
		return MicrosoftStoreExecutor{}
	}
	return nil
}
