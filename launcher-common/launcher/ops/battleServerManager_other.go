//go:build !darwin

package ops

import "github.com/luskaner/ageLANServer/common/game/executor/base"

func (c *Config) BattleServerRequired(_ base.Executor) bool {
	return c.gameRequiresBattleServer()
}
