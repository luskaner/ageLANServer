package main

import (
	"agent/internal"
	"fmt"
	launcherCommon "launcher-common"
	"launcher-common/executor"
	"os"
)

func main() {
	if !executor.IsAdmin() {
		fmt.Println("This proram must be run as an administrator")
		os.Exit(launcherCommon.ErrNotAdmin)
	}
	internal.RunIpcServer()
}
