package main

import (
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/launcher-common/cmdlog"
	"github.com/luskaner/ageLANServer/server-genCert/internal/cmd"
)

var version = "development"

func main() {
	cmdlog.Initialize()
	cmdlog.Banner(executables.ServerGenCert, version)
	cmd.Version = version
	common.ChdirToExe()
	err, exitCode := cmd.Execute()
	if err != nil {
		fmt.Print(err)
	}
	os.Exit(exitCode)
}
