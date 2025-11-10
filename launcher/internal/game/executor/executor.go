package executor

import (
	"github.com/luskaner/ageLANServer/common"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/game/appx"
	"github.com/luskaner/ageLANServer/common/game/steam"
)

type Exec interface {
	Do(args []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result)
	GameProcesses() (steamProcess bool, xboxProcess bool)
}

type baseExec struct {
	gameId string
}

type SteamExec struct {
	baseExec
	libraryFolder string
}
type XboxExec struct {
	baseExec
	gamePath string
}
type CustomExec struct {
	Executable string
}

func (exec SteamExec) Do(_ []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	return startUri(steam.NewGame(exec.gameId).OpenUri(), optionsFn)
}

func (exec SteamExec) GameProcesses() (steamProcess bool, xboxProcess bool) {
	steamProcess = true
	return
}

func (exec SteamExec) GamePath() string {
	return steam.NewGame(exec.gameId).Path(exec.libraryFolder)
}

func (exec CustomExec) execute(args []string, admin bool, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
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

func (exec CustomExec) Do(args []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	result = exec.execute(args, false, optionsFn)
	return
}

func (exec CustomExec) DoElevated(args []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	result = exec.execute(args, true, optionsFn)
	return
}

func steamExec(gameId string) (ok bool, executor Exec) {
	steamGame := steam.NewGame(gameId)
	if libraryFolder := steamGame.LibraryFolder(); libraryFolder != "" {
		ok = true
		executor = SteamExec{baseExec{gameId: gameId}, libraryFolder}
	}
	return
}

func xboxExec(gameId string) (ok bool, executor Exec) {
	if gameId != common.GameAoM {
		var gameLocation string
		if ok, gameLocation = appx.GameInstallLocation(gameId); ok {
			executor = XboxExec{baseExec{gameId: gameId}, gameLocation}
		}
	}
	return
}

func MakeExec(gameId string, executable string) Exec {
	if executable != "auto" {
		switch executable {
		case "steam":
			if ok, executor := steamExec(gameId); ok {
				return executor
			}
		case "msstore":
			if ok, executor := xboxExec(gameId); ok {
				return executor
			}
		default:
			return CustomExec{Executable: executable}
		}
		return nil
	}
	if ok, executor := steamExec(gameId); ok {
		return executor
	}
	if ok, executor := xboxExec(gameId); ok {
		return executor
	}
	return nil
}
