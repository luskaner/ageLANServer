package main

import (
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/logger"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/cmd"
)

var version = "development"

func main() {
	commonLogger.Initialize(os.Stdout)
	if !executor.IsAdmin() {
		commonLogger.Println("This program must be run as an administrator")
		os.Exit(launcherCommon.ErrNotAdmin)
	}
	common.ChdirToExe()
	cmd.Version = version
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
