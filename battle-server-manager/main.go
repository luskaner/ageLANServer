package main

import (
	"battle-server-manager/internal/cmd"
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/logger"
)

var version = "development"

func main() {
	commonLogger.Initialize(os.Stdout)
	cmd.Version = version
	common.ChdirToExe()
	err, exitCode := cmd.Execute()
	if err != nil {
		fmt.Print(err)
	}
	os.Exit(exitCode)
}
