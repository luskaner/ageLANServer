package cmdUtils

import (
	"github.com/luskaner/ageLANServer/common/game/executor/base"
	"github.com/luskaner/ageLANServer/common/game/executor/steam"
	commonSteam "github.com/luskaner/ageLANServer/common/game/steam"
)

func (c *Config) IsolationPath(executer base.Executor) (path string) {
	switch executer.(type) {
	case *steam.Exec:
		return commonSteam.UserProfilePath(c.gameId)
	default:
		return c.wineIsolationPath(executer)
	}
}
