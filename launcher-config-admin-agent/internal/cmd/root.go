package cmd

import (
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/fileLock"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/config"
	launcherCommonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor"
	"github.com/luskaner/ageLANServer/launcher-config-admin-agent/internal"
	"github.com/luskaner/ageLANServer/launcher-config-admin-agent/internal/ipc"
	"github.com/spf13/pflag"
)

var (
	Version string
	values  *config.FlushCacheValues
)

func Execute() error {
	var singleFs *cmd.SingleFlagSet
	values, singleFs = config.FlushCacheSingleFlagSet(Version, runRoot)
	return singleFs.Execute()
}

func runRoot(_ *pflag.FlagSet) error {
	commonLogger.Initialize(nil)
	if values.LogRoot != "" {
		internal.InitializeOrExit(values.LogRoot)
	}
	lock := &fileLock.PidLock{}
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
		return nil
	}
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
	if values.IPs || values.Certs {
		if values.IPs {
			commonLogger.Println("Flushing IP cache...")
		}
		if values.Certs {
			commonLogger.Println("Flushing certificate cache...")
		}
		var result *exec.Result
		if buffErr := commonLogger.FileLogger.Buffer("config-admin_flushCache", func(writer io.Writer) {
			_, result = launcherCommonExecutor.RunFlushCache(values.IPs, values.Certs, values.LogRoot, writer, func(options exec.Options) {
				if writer != nil {
					commonLogger.Println("run config admin flushCache", options.String())
				}
			})
		}); buffErr != nil {
			exitCode = common.ErrFileLog
			return nil
		}
		if !result.Success() {
			commonLogger.Println("Failed to flush cache with exit code: ", result.ExitCode)
			if result.Err != nil {
				commonLogger.Println(result.Err.Error())
			}
			exitCode = internal.ErrFlushCache
			return nil
		}
	}
	exitCode = ipc.StartServer(values.LogRoot)
	return nil
}
