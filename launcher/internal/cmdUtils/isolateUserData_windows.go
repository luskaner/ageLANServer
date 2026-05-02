package cmdUtils

import (
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/game/executor/base"
)

func (c *Config) IsolationPath(_ base.Executor) (path string) {
	return game.UserProfilePath(c.gameId)
}
