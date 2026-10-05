package main

import (
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/launcher-agent/internal/cmd"
	"github.com/luskaner/ageLANServer/launcher-common/cmdlog"
)

var version = "development"

func main() {
	// The agent normally runs windowless in the background, but everything it has
	// to say is read by whoever launched it, so it prints like everything else.
	cmdlog.Initialize()
	cmdlog.Banner(executables.LauncherAgent, version)
	cmd.Version = version
	common.ChdirToExe()
	err, exitCode := cmd.Execute()
	if err != nil {
		fmt.Print(err)
	}
	os.Exit(exitCode)
}
