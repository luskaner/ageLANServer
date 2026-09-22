package cmd

import (
	"io"

	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/fileLock"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	launcherCommonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor"
	"github.com/luskaner/ageLANServer/launcher-config-admin-agent/internal"
	"github.com/luskaner/ageLANServer/launcher-config-admin-agent/internal/ipc"
)

var (
	isAdminFn          = executor.IsAdmin
	initializeOrExitFn = internal.InitializeOrExit
	runFlushCacheFn    = launcherCommonExecutor.RunFlushCache
	startServerFn      = ipc.StartServer
	newPidLockFn       = func() *fileLock.PidLock { return &fileLock.PidLock{} }
	pidLockFn          = func(l *fileLock.PidLock) error { return l.Lock() }
	pidUnlockFn        = func(l *fileLock.PidLock) error { return l.Unlock() }
	loggerInitFn       = commonLogger.Initialize
	loggerCloseFn      = commonLogger.CloseFileLog
	bufferFn           = func(name string, fn func(writer io.Writer)) error {
		if commonLogger.FileLogger == nil {
			fn(nil)
			return nil
		}
		return commonLogger.FileLogger.Buffer(name, fn)
	}
)
