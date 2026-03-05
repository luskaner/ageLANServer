package main

import (
	"battle-server-manager/internal/cmd"
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/logger"
)

var version = "development"

func main() {
	commonLogger.Initialize(os.Stdout)
	cmd.Version = version
	common.ChdirToExe()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
