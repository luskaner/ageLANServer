package main

import (
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/common/pidLock"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"github.com/luskaner/aoe2DELanServer/launcher-config-admin-agent/internal"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	lock := &pidLock.Lock{}
	if err := lock.Lock(); err != nil {
		os.Exit(common.ErrPidLock)
	}
	if !exec.IsAdmin() {
		_ = lock.Unlock()
		os.Exit(launcherCommon.ErrNotAdmin)
	}
	common.ChdirToExe()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			_ = lock.Unlock()
			os.Exit(common.ErrSignal)
		}
	}()
	errorCode := internal.RunIpcServer()
	_ = lock.Unlock()
	os.Exit(errorCode)
}
