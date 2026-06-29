package main

import (
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server-genCert/internal/cmd"
)

var version = "development"

func main() {
	cmd.Version = version
	common.ChdirToExe()
	err, exitCode := cmd.Execute()
	if err != nil {
		print(err)
	}
	os.Exit(exitCode)
}
