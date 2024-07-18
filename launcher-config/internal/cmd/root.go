package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "config",
	Short: "config execute config-only tasks",
	Long:  "config execute config-only tasks as required by launcher directly or indirectly via the watcher",
}

var Version string

func Execute() error {
	initSetUp()
	initRevert()
	rootCmd.PersistentFlags().StringVarP(&Version, "version", "v", Version, "Version")
	return rootCmd.Execute()
}
