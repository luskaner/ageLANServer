package main

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/cmd"
	"github.com/spf13/cobra"
)

const version = "development"

func main() {
	cobra.MousetrapHelpText = ""
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
	cmd.Version = version
}
