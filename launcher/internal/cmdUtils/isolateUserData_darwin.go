package cmdUtils

import (
	"github.com/luskaner/ageLANServer/common/game/executor/base"
)

func (c *Config) IsolationPath(executer base.Executor) (path string) {
	return c.wineIsolationPath(executer)
}
