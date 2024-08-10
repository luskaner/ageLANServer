package main

import (
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor"
	"github.com/luskaner/aoe2DELanServer/launcher-config-admin/internal"
	"github.com/luskaner/aoe2DELanServer/launcher-config-admin/internal/cmd"
	"os"
)

const version = "development"

func main() {
	if !executor.IsAdmin() {
		fmt.Println("This program must be run as an administrator")
		os.Exit(launcherCommon.ErrNotAdmin)
	}
	if !internal.ParentMatches(common.GetExeFileName(true, common.LauncherConfig)) {
		fmt.Printf("This program should only be run through \"%s\", not directly. You can use the same arguments and more.\n", common.LauncherConfig)
	}
	cmd.Version = version
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
