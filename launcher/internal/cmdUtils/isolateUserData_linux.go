package cmdUtils

import (
	"github.com/luskaner/ageLANServer/common/game/executor/base"
	"github.com/luskaner/ageLANServer/common/game/executor/steam"
	"github.com/luskaner/ageLANServer/common/game/executor/steam/wine"
	"github.com/luskaner/ageLANServer/common/game/executor/steam/wine/crossover"
	wineExec "github.com/luskaner/ageLANServer/common/game/executor/wine"
	commonSteam "github.com/luskaner/ageLANServer/common/game/steam"
	wineSteam "github.com/luskaner/ageLANServer/common/game/wine/steam"
)

func (c *Config) IsolationPath(executer base.Executor) (path string) {
	switch executer.(type) {
	case *steam.Exec:
		return commonSteam.UserProfilePath(c.gameId)
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
