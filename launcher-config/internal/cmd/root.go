package cmd

import (
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	commonUserData "github.com/luskaner/ageLANServer/launcher-common/userData"
)

var Version string
var path *commonUserData.Path
var rootFlagSet *commonCmd.RootFlagSet

func Execute() (err error, exitCode int) {
	rootFlagSet = commonCmd.NewRootFlagSet()
	rootFlagSet.RegisterCommand("setup", runSetUp)
	rootFlagSet.RegisterCommand("revert", runRevert)
	rootFlagSet.RegisterCommand("flushCache", runFlushCache)
	return rootFlagSet.Execute(Version)
}
