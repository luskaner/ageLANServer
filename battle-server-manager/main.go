package main

import (
	"battle-server-manager/internal/cmd"

	"github.com/luskaner/ageLANServer/common"
	"github.com/spf13/cobra"
)

const version = "development"

func main() {
	cobra.MousetrapHelpText = ""
	cmd.Version = version
	common.ChdirToExe()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
