package main

import (
	"battle-server-manager/internal/cmd"
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/launcher-common/cmdlog"
)

var version = "development"

func main() {
	cmdlog.Initialize()
	cmdlog.Banner(executables.BattleServerManager, version)
	cmd.Version = version
	common.ChdirToExe()
	err, exitCode := cmd.Execute()
	if err != nil {
		fmt.Print(err)
	}
	os.Exit(exitCode)
}
