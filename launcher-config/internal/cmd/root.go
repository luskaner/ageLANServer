package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var RootCmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "config execute config-only tasks",
	Long:  "config execute config-only tasks as required by launcher directly or indirectly via the agent",
}

var Version string

func Execute() error {
	RootCmd.Version = Version
	InitSetUp()
	InitRevert()
	return RootCmd.Execute()
}
