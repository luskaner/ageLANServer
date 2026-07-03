package main

import (
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-config/internal/cmd"
)

var version = "development"

func main() {
	commonLogger.Initialize(os.Stdout)
	if executor.IsAdmin() {
		commonLogger.Println("Running as administrator, this is not recommended for security reasons. It will request isolated admin privileges if/when it needs.")
	}
	common.ChdirToExe()
	cmd.Version = version
	err, exitCode := cmd.Execute()
	if err != nil {
		fmt.Print(err)
	}
	os.Exit(exitCode)
}
