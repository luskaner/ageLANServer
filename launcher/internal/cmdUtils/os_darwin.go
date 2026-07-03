package cmdUtils

import (
	"runtime"

	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/game/executor/base"
	"github.com/luskaner/ageLANServer/common/game/executor/custom"
	"github.com/luskaner/ageLANServer/common/game/executor/steam"
)

func (c *Config) NativeMacOsGame(executer base.Executor, considerCustomLauncher bool) bool {
	if c.gameId == game.AoE2 && runtime.GOARCH == "arm64" {
		if _, ok := executer.(steam.Exec); ok {
			return true
		}
		if considerCustomLauncher {
			if _, ok := executer.(custom.Exec); ok {
				return true
			}
		}
	}
	return false
}
