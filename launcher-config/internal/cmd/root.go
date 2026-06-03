package cmd

import (
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	commonUserData "github.com/luskaner/ageLANServer/launcher-common/userData"
)

var Version string
var path *commonUserData.Path
var rootFlagSet *commonCmd.RootFlagSet
var errorCode int

func Execute() error {
	rootFlagSet = commonCmd.NewRootFlagSet()
	rootFlagSet.RegisterCommand("setup", runSetUp)
	rootFlagSet.RegisterCommand("revert", runRevert)
	return rootFlagSet.Execute(Version)
}
