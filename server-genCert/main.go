package main

import (
	"github.com/luskaner/aoe2DELanServer/server-genCert/internal/cmd"
	"github.com/spf13/cobra"
)

const version = "development"

func main() {
	cobra.MousetrapHelpText = ""
	cmd.Version = version
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
