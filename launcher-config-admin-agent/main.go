package main

import (
	"agent/internal"
	"common"
	"common/pidLock"
	"fmt"
	launcherCommon "launcher-common"
	"launcher-common/executor"
	"os"
)

func main() {
	lock := &pidLock.Lock{}
	if err := lock.Lock(); err != nil {
		fmt.Println("Failed to lock pid file. You may try checking if the process in PID file exists (and killing it).")
		os.Exit(common.ErrPidLock)
	}
	if !executor.IsAdmin() {
		fmt.Println("This program must be run as an administrator")
		_ = lock.Unlock()
		os.Exit(launcherCommon.ErrNotAdmin)
	}
	errorCode := internal.RunIpcServer()
	_ = lock.Unlock()
	os.Exit(errorCode)
}
