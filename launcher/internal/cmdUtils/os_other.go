//go:build !darwin

package cmdUtils

import "github.com/luskaner/ageLANServer/common/game/executor/base"

func (c *Config) NativeMacOsGame(_ base.Executor, _ bool) bool {
	return false
}
