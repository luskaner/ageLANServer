package game

import (
	"github.com/luskaner/ageLANServer/common"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/game/appx"
	"github.com/luskaner/ageLANServer/common/game/steam"
)

type Executor interface {
	Execute(args []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result)
	GameProcesses() (steamProcess bool, xboxProcess bool)
}

type baseExecutor struct {
	gameId string
}

type SteamExecutor struct {
	baseExecutor
	libraryFolder string
}
type XboxExecutor struct {
	baseExecutor
	gamePath string
}
type CustomExecutor struct {
	Executable string
}

func (exec SteamExecutor) Execute(_ []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	return startUri(steam.NewGame(exec.gameId).OpenUri(), optionsFn)
}

func (exec SteamExecutor) GameProcesses() (steamProcess bool, xboxProcess bool) {
	steamProcess = true
	return
}

func (exec SteamExecutor) GamePath() string {
	return steam.NewGame(exec.gameId).Path(exec.libraryFolder)
}

func (exec CustomExecutor) execute(args []string, admin bool, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	options := commonExecutor.Options{File: exec.Executable, Args: args}
	if admin {
		options.AsAdmin = true
	}
	options.ShowWindow = true
	options.GUI = true
	optionsFn(options)
	result = options.Exec()
	return
}

func (exec CustomExecutor) Execute(args []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	result = exec.execute(args, false, optionsFn)
	return
}

func (exec CustomExecutor) ExecuteElevated(args []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	result = exec.execute(args, true, optionsFn)
	return
}

func steamExecutor(gameId string) (ok bool, executor Executor) {
	steamGame := steam.NewGame(gameId)
	if libraryFolder := steamGame.LibraryFolder(); libraryFolder != "" {
		ok = true
		executor = SteamExecutor{baseExecutor{gameId: gameId}, libraryFolder}
	}
	return
}

func xboxExecutor(gameId string) (ok bool, executor Executor) {
	if gameId != common.GameAoM {
		var gameLocation string
		if ok, gameLocation = appx.GameInstallLocation(gameId); ok {
			executor = XboxExecutor{baseExecutor{gameId: gameId}, gameLocation}
		}
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
			if ok, executor := xboxExecutor(gameId); ok {
				return executor
			}
		default:
			return CustomExecutor{Executable: executable}
		}
		return nil
	}
	if ok, executor := steamExecutor(gameId); ok {
		return executor
	}
	if ok, executor := xboxExecutor(gameId); ok {
		return executor
	}
	return nil
}
