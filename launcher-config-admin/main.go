package main

import (
	"admin/internal"
	"admin/internal/cmd"
	"common"
	"fmt"
	launcherCommon "launcher-common"
	"launcher-common/executor"
	"os"
)

const version = "development"

func main() {
	if !executor.IsAdmin() {
		fmt.Println("This program must be run as an administrator")
		os.Exit(launcherCommon.ErrNotAdmin)
	}
	if !internal.ParentMatches(common.GetExeFileName(common.LauncherConfig)) {
		fmt.Printf("This program should only be run through \"%s\", not directly. You can use the same arguments and more.\n", common.LauncherConfig)
	}
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
	cmd.Version = version
}
