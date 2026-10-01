package cmd

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/fileLock"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-agent/internal"
	"github.com/luskaner/ageLANServer/launcher-agent/internal/watch"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/agent"
	"github.com/spf13/pflag"
)

var Version string
var values *agent.Values

var (
	createLockFn   = func() fileLock.Locker { return &fileLock.PidLock{} }
	chdirToExeFn   = common.ChdirToExe
	initializeFn   = internal.Initialize
	watchFn        = watch.Watch
	signalNotifyFn = signal.Notify
)

func Execute() (err error, exitCode int) {
	var singleFs *cmd.SingleFlagSet
	values, singleFs = agent.SingleFlagSet(Version, runRoot)
	return singleFs.Execute()
}

func runRoot(_ *pflag.FlagSet) (err error, exitCode int) {
	commonLogger.Initialize(os.Stdout)
	lock := createLockFn()
	if err = lock.Lock(); err != nil {
		commonLogger.Println("Failed to lock pid file. Kill process 'agent' if it is running in your task manager.")
		exitCode = common.ErrPidLock
		return
	}
	chdirToExeFn()
	if values.LogRoot != "" && values.BaseDataPath != "" {
		initializeFn(values.LogRoot)
	}
	var cleanupOnce sync.Once
	sigs := make(chan os.Signal, 1)
	signalNotifyFn(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			commonLogger.Println("Received terminate signal, shutting down...")
			exitCode = common.ErrSignal
			cleanupOnce.Do(func() {
				watch.Cleanup(values, &exitCode)
			})
			if err = lock.Unlock(); err != nil {
				commonLogger.Printf("Failed to unlock: %v\n", err)
			}
			commonLogger.Printf("Exit code: %d\n", exitCode)
		}
	}()
	watchFn(
		values,
		&exitCode,
		&cleanupOnce,
	)
	_ = lock.Unlock()
	return
}
