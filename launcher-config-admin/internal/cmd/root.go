package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var rootCmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "config-admin execute admin-only tasks",
	Long:  "config-admin execute admin-only tasks as required by config",
}

var Version string

func Execute() error {
	rootCmd.Version = Version
	initSetUp()
	initRevert()
	return rootCmd.Execute()
}
