package main

import (
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/pidLock"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-config-admin-agent/internal"
)

func main() {
	commonLogger.Initialize(nil)
	logRoot := os.Args[1]
	if logRoot != "-" {
		internal.InitializeOrExit(logRoot)
	} else {
		logRoot = ""
	}
	lock := &pidLock.Lock{}
	if err := lock.Lock(); err != nil {
		commonLogger.Println("Failed to lock pid file. Kill process 'config-admin-agent' if it is running in your task manager.")
		commonLogger.CloseFileLog()
		os.Exit(common.ErrPidLock)
	}
	exitCode := common.ErrSuccess
	defer func() {
		commonLogger.CloseFileLog()
		if r := recover(); r != nil {
			commonLogger.Println(r)
			commonLogger.Println(string(debug.Stack()))
			exitCode = common.ErrGeneral
		}
		_ = lock.Unlock()
		os.Exit(exitCode)
	}()
	if !executor.IsAdmin() {
		commonLogger.Println("Program must be run as admin")
		exitCode = launcherCommon.ErrNotAdmin
		return
	}
	common.ChdirToExe()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			commonLogger.CloseFileLog()
			_ = lock.Unlock()
			os.Exit(common.ErrSignal)
		}
	}()
	exitCode = internal.RunIpcServer(logRoot)
}
