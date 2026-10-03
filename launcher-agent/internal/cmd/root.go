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
	"github.com/luskaner/ageLANServer/launcher-common/cmdlog"
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
		cmdlog.Fail("Failed to lock pid file. Kill process agent if it is running in your task manager.")
		exitCode = common.ErrPidLock
		return
	}
	chdirToExeFn()
	if values.LogRoot != "" && values.BaseDataPath != "" {
		initializeFn(values.LogRoot)
	}
	var cleanupOnce sync.Once
	// Shared by the signal handler below and Watch, which run concurrently, so
	// the exit code goes through a synchronised holder rather than a bare *int.
	code := watch.NewExitCode()
	sigs := make(chan os.Signal, 1)
	signalNotifyFn(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			cmdlog.Step("Received terminate signal, shutting down...")
			code.Set(common.ErrSignal)
			cleanupOnce.Do(func() {
				watch.Cleanup(values, code)
			})
			if err = lock.Unlock(); err != nil {
				cmdlog.Fail("Failed to unlock: %s", err)
			}
			cmdlog.Fault("Exit code: %d", code.Get())
		}
	}()
	watchFn(
		values,
		code,
		&cleanupOnce,
	)
	exitCode = code.Get()
	_ = lock.Unlock()
	return
}
