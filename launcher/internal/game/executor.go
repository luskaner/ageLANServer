package game

import (
	"github.com/luskaner/ageLANServer/common"
	commonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher-common/steam"
	"os"
)

type Executor interface {
	Execute(args []string) (result *commonExecutor.Result)
	GameProcesses() (steamProcess bool, xboxProcess bool)
}

type baseExecutor struct {
	gameTitle common.GameTitle
}

type SteamExecutor struct {
	baseExecutor
}
type XboxExecutor struct {
	baseExecutor
}
type CustomExecutor struct {
	Executable string
}

func (exec SteamExecutor) Execute(_ []string) (result *commonExecutor.Result) {
	return startUri(steam.NewGame(exec.gameTitle).OpenUri())
}

func (exec SteamExecutor) GameProcesses() (steamProcess bool, xboxProcess bool) {
	steamProcess = true
	return
}

func (exec CustomExecutor) execute(args []string, admin bool) (result *commonExecutor.Result) {
	options := commonExecutor.Options{File: exec.Executable, Args: args}
	if admin {
		options.AsAdmin = true
		options.ShowWindow = true
	}
	result = options.Exec()
	return
}

func (exec CustomExecutor) Execute(args []string) (result *commonExecutor.Result) {
	result = exec.execute(args, false)
	return
}

func (exec CustomExecutor) ExecuteElevated(args []string) (result *commonExecutor.Result) {
	result = exec.execute(args, true)
	return
}

func isInstalledCustom(executable string) bool {
	info, err := os.Stat(executable)
	if err != nil || os.IsNotExist(err) || info.IsDir() {
		return false
	}
	return true
}

func steamExecutor(gameTitle common.GameTitle) (ok bool, executor Executor) {
	steamGame := steam.NewGame(gameTitle)
	if steamGame.GameInstalled() {
		ok = true
		executor = SteamExecutor{baseExecutor{gameTitle: gameTitle}}
	}
	return
}

func xboxExecutor(gameTitle common.GameTitle) (ok bool, executor Executor) {
	if isInstalledOnXbox(gameTitle) {
		ok = true
		executor = XboxExecutor{baseExecutor{gameTitle: gameTitle}}
	}
	return
}

func MakeExecutor(gameTitle common.GameTitle, launcher common.ClientLauncher, executable string) Executor {
	if launcher != common.ClientLauncherSteamOrMSStore {
		switch launcher {
		case common.ClientLauncherSteam:
			if ok, executor := steamExecutor(gameTitle); ok {
				return executor
			}
		case common.ClientLauncherMSStore:
			if ok, executor := xboxExecutor(gameTitle); ok {
				return executor
			}
		case common.ClientLauncherPath:
			if isInstalledCustom(executable) {
				return CustomExecutor{Executable: executable}
			}
		}
		return nil
	}
	if ok, executor := steamExecutor(gameTitle); ok {
		return executor
	}
	if ok, executor := xboxExecutor(gameTitle); ok {
		return executor
	}
	return nil
}
