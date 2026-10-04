//go:build !darwin

package cmdUtils

import "github.com/luskaner/ageLANServer/common/game/executor/base"

func (c *Config) BattleServerRequired(_ base.Executor) bool {
	return c.gameRequiresBattleServer()
}
