package main

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server-genCert/internal/cmd"
)

var version = "development"

func main() {
	cmd.Version = version
	common.ChdirToExe()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
