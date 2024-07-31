package main

import (
	"cfgAdminAgent/internal"
	"common"
	"common/pidLock"
	launcherCommon "launcherCommon"
	"launcherCommon/executor"
	"os"
)

func main() {
	lock := &pidLock.Lock{}
	if err := lock.Lock(); err != nil {
		os.Exit(common.ErrPidLock)
	}
	if !executor.IsAdmin() {
		_ = lock.Unlock()
		os.Exit(launcherCommon.ErrNotAdmin)
	}
	errorCode := internal.RunIpcServer()
	_ = lock.Unlock()
	os.Exit(errorCode)
}
