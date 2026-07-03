package main

import (
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-config-admin-agent/internal/cmd"
)

var version = "development"

func main() {
	cmd.Version = version
	common.ChdirToExe()
	err, exitCode := cmd.Execute()
	if err != nil {
		fmt.Print(err)
	}
	os.Exit(exitCode)
}
