package main

import (
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/logger"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/cmd"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/parentCheck"
)

const version = "development"

func main() {
	commonLogger.Initialize(os.Stdout)
	if !executor.IsAdmin() {
		commonLogger.Println("This program must be run as an administrator")
		os.Exit(launcherCommon.ErrNotAdmin)
	}
	// FIXME: Always returns false
	if !parentCheck.ParentMatches() {
		commonLogger.Printf("This program should only be run through \"%s\", not directly. You can use the same arguments and more.\n", common.LauncherConfig)
	}
	common.ChdirToExe()
	cmd.Version = version
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
