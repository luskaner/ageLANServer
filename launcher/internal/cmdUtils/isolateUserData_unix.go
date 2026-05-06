//go:build !windows

package cmdUtils

import (
	"github.com/luskaner/ageLANServer/common/game/executor/base"
	"github.com/luskaner/ageLANServer/common/game/executor/steam/wine"
	"github.com/luskaner/ageLANServer/common/game/executor/steam/wine/crossover"
	wineExec "github.com/luskaner/ageLANServer/common/game/executor/wine"
	wineSteam "github.com/luskaner/ageLANServer/common/game/wine/steam"
)

func (c *Config) wineIsolationPath(executer base.Executor) (path string) {
	switch executer.(type) {
	case *wine.Exec, *crossover.Exec:
		return wineSteam.OutputLauncherConfigHelper(
			executer.(wineExec.CustomExec),
			"userProfilePath",
			[]string{
				c.gameId,
			},
		)
	}
	return ""
}
