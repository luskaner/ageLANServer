package cmdUtils

import (
	"github.com/luskaner/ageLANServer/common/game/executor/base"
)

func (c *Config) BattleServerRequired(executer base.Executor) bool {
	return c.gameRequiresBattleServer() || c.NativeMacOsGame(executer, true)
}
