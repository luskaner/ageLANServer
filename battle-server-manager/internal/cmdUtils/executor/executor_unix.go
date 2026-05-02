//go:build !windows

package executor

import (
	"fmt"

	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/game/executor/steam/wine"
	"github.com/luskaner/ageLANServer/common/game/executor/steam/wine/crossover"
	wineExec "github.com/luskaner/ageLANServer/common/game/executor/wine"
)

type LocatableString interface {
	game.Locatable
	fmt.Stringer
}

func steamCrossOverPath(gameId string, battleServerPath string) (path string) {
	if ex, ok := wine.NewExec(gameId); ok {
		path = winePath(ex, gameId, battleServerPath)
	}
	return
}

func steamWinePath(gameId string, battleServerPath string) (path string) {
	if ex, ok := crossover.NewExec(gameId); ok {
		path = winePath(ex, gameId, battleServerPath)
	}
	return
}

func winePath(loc LocatableString, gameId string, battleServerPath string) (path string) {
	return locatablePath(func(gameId string) (game game.Locatable, ok bool) {
		game = loc
		ok = true
		return
	}, gameId, battleServerPath, loc.String())
}

func execWithOptions(gameId string, options *exec.Options) exec.Result {
	var foundExecutor wineExec.CustomExec
	if executor, ok := wineExec.NewExec(); ok {
		foundExecutor = executor
	} else if executor, ok := crossover.NewExec(gameId); ok {
		foundExecutor = executor
	}
	if foundExecutor != nil {
		args := []string{options.File}
		args = append(args, options.Args...)
		result := foundExecutor.DoCustom(
			args,
			func(currentOptions *exec.Options) {
				currentOptions.UseWorkingPath = options.UseWorkingPath
				currentOptions.Wait = options.Wait
				currentOptions.Stdout = options.Stdout
				currentOptions.Stderr = options.Stderr
				currentOptions.Pid = options.Pid
				currentOptions.ExitCode = options.ExitCode
				currentOptions.ShowWindow = options.ShowWindow
			},
		)
		return *result
	}
	return exec.Result{
		Err: fmt.Errorf("no suitable executor found for game %s", gameId),
	}
}
