//go:build !darwin

package cmdUtils

import "github.com/luskaner/ageLANServer/common/game/executor/base"

func (c *Config) GamePathToGameCertPath(_ base.Executor, path string) string {
	return path
}
