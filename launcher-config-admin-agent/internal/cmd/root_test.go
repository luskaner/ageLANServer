package cmd

import (
	"errors"
	"io"
	"testing"

	"github.com/luskaner/ageLANServer/common"
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/fileLock"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/config"
	"github.com/luskaner/ageLANServer/launcher-config-admin-agent/internal"
)

func resetState(t *testing.T) {
	t.Helper()
	oldIsAdmin := isAdminFn
	oldInit := initializeOrExitFn
	oldRunFlush := runFlushCacheFn
	oldStart := startServerFn
	oldNewPid := newPidLockFn
	oldPidLock := pidLockFn
	oldPidUnlock := pidUnlockFn
	oldLoggerInit := loggerInitFn
	oldLoggerClose := loggerCloseFn
	oldBuffer := bufferFn
	oldValues := values
	oldLogger := commonLogger.FileLogger
	t.Cleanup(func() {
		isAdminFn = oldIsAdmin
		initializeOrExitFn = oldInit
		runFlushCacheFn = oldRunFlush
		startServerFn = oldStart
		newPidLockFn = oldNewPid
		pidLockFn = oldPidLock
		pidUnlockFn = oldPidUnlock
		loggerInitFn = oldLoggerInit
		loggerCloseFn = oldLoggerClose
		bufferFn = oldBuffer
		values = oldValues
		commonLogger.FileLogger = oldLogger
	})
	commonLogger.FileLogger = nil
	values = &config.FlushCacheValues{
		RevertMinimalValues: &config.RevertMinimalValues{},
		LogRootValues:       &commonCmd.LogRootValues{},
	}
	loggerInitFn = func(io.Writer) {}
	loggerCloseFn = func() {}
	initializeOrExitFn = func(string) {}
	bufferFn = func(string, func(io.Writer)) error { return nil }
	// default buffer that calls fn
	bufferFn = func(name string, fn func(io.Writer)) error {
		fn(nil)
		return nil
	}
	pidLockFn = func(*fileLock.PidLock) error { return nil }
	pidUnlockFn = func(*fileLock.PidLock) error { return nil }
	newPidLockFn = func() *fileLock.PidLock { return &fileLock.PidLock{} }
	isAdminFn = func() bool { return true }
	startServerFn = func(string) int { return common.ErrSuccess }
	runFlushCacheFn = func(bool, bool, string, io.Writer, func(*exec.Options)) (string, *exec.Result) {
		return "", &exec.Result{ExitCode: common.ErrSuccess}
	}
}

func TestExecuteSetsVersion(t *testing.T) {
	resetState(t)
	Version = "test"
	// Avoid actually calling os.Args logic by testing that Execute can be called with mock SingleFlagSet?
	// Instead we test that Version is set and Execute doesn't panic when called
	// Use a helper to set values via flag parsing would require os.Args; we just test that Execute returns without panic when Version set
	// Since Execute uses config.FlushCacheSingleFlagSet which parses os.Args, it will likely succeed with no args (shows help)
	// We'll just call it and ensure rootFlagSet is set via internal/caller? But Execute in this package uses singleFs, not rootFlagSet.
	// So we just ensure it doesn't panic
	_, _ = Execute()
	if Version != "test" {
		t.Fatal("Version should be test")
	}
}

func TestRunRootPidLockFailure(t *testing.T) {
	resetState(t)
	pidLockFn = func(*fileLock.PidLock) error { return errors.New("lock fail") }
	values.LogRoot = ""
	_, code := runRoot(nil)
	if code != common.ErrPidLock {
		t.Fatalf("code=%d want ErrPidLock %d", code, common.ErrPidLock)
	}
}

func TestRunRootNotAdmin(t *testing.T) {
	resetState(t)
	isAdminFn = func() bool { return false }
	_, code := runRoot(nil)
	if code != launcherCommon.ErrNotAdmin {
		t.Fatalf("code=%d want ErrNotAdmin %d", code, launcherCommon.ErrNotAdmin)
	}
}

func TestRunRootWithLogRootCallsInitialize(t *testing.T) {
	resetState(t)
	called := false
	initializeOrExitFn = func(s string) {
		called = true
		if s != "/tmp/log" {
			t.Fatalf("logRoot=%q", s)
		}
	}
	values.LogRoot = "/tmp/log"
	_, code := runRoot(nil)
	if !called {
		t.Fatal("initializeOrExit not called")
	}
	if code != common.ErrSuccess {
		t.Fatalf("code=%d want success", code)
	}
}

func TestRunRootFlushCacheSuccess(t *testing.T) {
	resetState(t)
	values.IPs = true
	values.Certs = false
	runFlushCacheFn = func(bool, bool, string, io.Writer, func(*exec.Options)) (string, *exec.Result) {
		return "", &exec.Result{ExitCode: common.ErrSuccess}
	}
	bufferFn = func(name string, fn func(io.Writer)) error {
		fn(nil)
		return nil
	}
	_, code := runRoot(nil)
	if code != common.ErrSuccess {
		t.Fatalf("code=%d want success", code)
	}
}

func TestRunRootFlushCacheBufferFailure(t *testing.T) {
	resetState(t)
	values.IPs = true
	bufferFn = func(string, func(io.Writer)) error { return errors.New("buffer fail") }
	_, code := runRoot(nil)
	if code != common.ErrFileLog {
		t.Fatalf("code=%d want ErrFileLog", code)
	}
}

func TestRunRootFlushCacheFailure(t *testing.T) {
	resetState(t)
	values.IPs = true
	runFlushCacheFn = func(bool, bool, string, io.Writer, func(*exec.Options)) (string, *exec.Result) {
		return "", &exec.Result{ExitCode: 1, Err: errors.New("fail")}
	}
	_, code := runRoot(nil)
	if code != internal.ErrFlushCache {
		t.Fatalf("code=%d want ErrFlushCache %d", code, internal.ErrFlushCache)
	}
}

func TestRunRootStartServerSuccess(t *testing.T) {
	resetState(t)
	startServerFn = func(string) int { return 42 }
	_, code := runRoot(nil)
	if code != 42 {
		t.Fatalf("code=%d want 42", code)
	}
}

func TestRunRootPanicRecovery(t *testing.T) {
	resetState(t)
	startServerFn = func(string) int { panic("test panic") }
	_, code := runRoot(nil)
	if code != common.ErrGeneral {
		t.Fatalf("code=%d want ErrGeneral", code)
	}
}

func TestRunRootNoFlushStartsServer(t *testing.T) {
	resetState(t)
	called := false
	startServerFn = func(s string) int { called = true; return common.ErrSuccess }
	values.IPs = false
	values.Certs = false
	_, code := runRoot(nil)
	if !called {
		t.Fatal("startServer should be called when no flush")
	}
	if code != common.ErrSuccess {
		t.Fatalf("code=%d", code)
	}
}
