package main

import (
	"github.com/spf13/cobra"
	"server/internal/cmd"
)

const version = "development"

func main() {
	cobra.MousetrapHelpText = ""
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
	cmd.Version = version
}
