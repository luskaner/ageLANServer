package main

import (
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/common/pidLock"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor"
	"github.com/luskaner/aoe2DELanServer/launcher-config-admin-agent/internal"
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
