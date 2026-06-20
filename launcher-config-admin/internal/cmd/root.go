package cmd

import (
	"github.com/luskaner/ageLANServer/common/cmd"
)

var Version string
var rootFlagSet *cmd.RootFlagSet

func Execute() error {
	rootFlagSet = cmd.NewRootFlagSet()
	rootFlagSet.RegisterCommand("setup", runSetUp)
	rootFlagSet.RegisterCommand("revert", runRevert)
	rootFlagSet.RegisterCommand("flushCache", runFlushCache)
	return rootFlagSet.Execute(Version)
}
