package main

import (
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/cmd"
)

var version = "development"

func main() {
	commonLogger.Initialize(nil)
	cmd.Version = version
	common.ChdirToExe()
	err, exitCode := cmd.Execute()
	if err != nil {
		fmt.Print(err)
	}
	os.Exit(exitCode)
}
