package main

import (
	"battle-server-manager/internal/cmd"
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/spf13/cobra"
)

const version = "development"

func main() {
	commonLogger.Initialize(os.Stdout)
	cobra.MousetrapHelpText = ""
	cmd.Version = version
	common.ChdirToExe()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
