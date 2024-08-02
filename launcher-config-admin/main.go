package main

import (
	"fmt"
	"github.com/luskaner/aoe2DELanServer/cfgAdmin/internal"
	"github.com/luskaner/aoe2DELanServer/cfgAdmin/internal/cmd"
	"github.com/luskaner/aoe2DELanServer/common"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcherCommon"
	"github.com/luskaner/aoe2DELanServer/launcherCommon/executor"
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
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
	cmd.Version = version
}
