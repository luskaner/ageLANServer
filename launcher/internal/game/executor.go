package game

import (
	commonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher-common/steam"
	"os"
)

type Executor interface {
	Execute(args []string) (result *commonExecutor.Result)
	GameProcesses() (steamProcess bool, microsoftStoreProcess bool)
}

type baseExecutor struct {
	gameId string
}

type SteamExecutor struct {
	baseExecutor
}
type MicrosoftStoreExecutor struct {
	baseExecutor
}
type CustomExecutor struct {
	Executable string
}

func (exec SteamExecutor) Execute(_ []string) (result *commonExecutor.Result) {
	return startUri(steam.NewGame(exec.gameId).OpenUri())
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

func steamExecutor(gameId string) (ok bool, executor Executor) {
	steamGame := steam.NewGame(gameId)
	if steamGame.GameInstalled() {
		ok = true
		executor = SteamExecutor{baseExecutor{gameId: gameId}}
	}
	return
}

func microsoftStoreExecutor(gameId string) (ok bool, executor Executor) {
	if isInstalledOnMicrosoftStore(gameId) {
		ok = true
		executor = MicrosoftStoreExecutor{baseExecutor{gameId: gameId}}
	}
	return
}

func MakeExecutor(gameId string, executable string) Executor {
	if executable != "auto" {
		switch executable {
		case "steam":
			if ok, executor := steamExecutor(gameId); ok {
				return executor
			}
		case "msstore":
			if ok, executor := microsoftStoreExecutor(gameId); ok {
				return executor
			}
		default:
			if isInstalledCustom(executable) {
				return CustomExecutor{Executable: executable}
			}
		}
		return nil
	}
	if ok, executor := steamExecutor(gameId); ok {
		return executor
	}
	if ok, executor := microsoftStoreExecutor(gameId); ok {
		return executor
	}
	return nil
}
