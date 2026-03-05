package cmd

import (
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/spf13/pflag"
)

var gamePath string
var Version string
var hostFilePath string
var certFilePath string
var logRoot string
var rootFlagSet *commonCmd.RootFlagSet

func Execute() error {
	rootFlagSet = commonCmd.NewRootFlagSet()
	rootFlagSet.RegisterCommand("setup", runSetUp)
	rootFlagSet.RegisterCommand("revert", runRevert)
	return rootFlagSet.Execute(Version)
}

func addGamePathFlag(fs *pflag.FlagSet) {
	fs.StringVar(
		&gamePath,
		"gamePath",
		"",
		"Path to the game folder. Required when using 'caStoreCert' and all except AoE: DE and AoE IV: AE.",
	)
}
