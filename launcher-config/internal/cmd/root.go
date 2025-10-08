package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "config execute config-only tasks",
	Long:  "config execute config-only tasks as required by 'launcher' directly or indirectly via the 'agent'",
}

var gamePath string
var Version string
var hostFilePath string
var certFilePath string

func Execute() error {
	RootCmd.Version = Version
	InitSetUp()
	InitRevert()
	return RootCmd.Execute()
}

func addGamePathFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&gamePath,
		"gamePath",
		"",
		"Path to the game folder. Required when using 'caStoreCert' and all except AoE I: DE..",
	)
}
