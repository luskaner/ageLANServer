package cmd

import (
	"github.com/luskaner/ageLANServer/common/cmd"
)

var Version string
var logRoot string
var rootFlagSet *cmd.RootFlagSet

func Execute() error {
	rootFlagSet = cmd.NewRootFlagSet()
	rootFlagSet.RegisterCommand("setup", runSetUp)
	rootFlagSet.RegisterCommand("revert", runRevert)
	return rootFlagSet.Execute(Version)
}
