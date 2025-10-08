package main

import (
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/cmd"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/parentCheck"
)

const version = "development"

func main() {
	if !executor.IsAdmin() {
		fmt.Println("This program must be run as an administrator")
		os.Exit(launcherCommon.ErrNotAdmin)
	}
	if !parentCheck.ParentMatches() {
		fmt.Printf("This program should only be run through \"%s\", not directly. You can use the same arguments and more.\n", common.LauncherConfig)
	}
	common.ChdirToExe()
	cmd.Version = version
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
