package cmd

import (
	"errors"
	"io"
	"testing"

	"github.com/luskaner/ageLANServer/common"
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/cmd/server"
	"github.com/luskaner/ageLANServer/common/fileLock"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/server/internal"
	"github.com/spf13/pflag"
)

func resetState(t *testing.T) {
	t.Helper()
	oldNewPid := fileLockNewFn
	oldLock := fileLockLockFn
	oldUnlock := fileLockUnlockFn
	oldLoggerInit := commonLoggerInitFn
	oldLoggerClose := commonLoggerCloseFn
	oldOpenLog := loggerOpenMainFileLogFn
	oldIsAdmin := isAdminFn
	oldInitCfg := initConfigFn
	oldValues := values
	oldConnectivity := internal.Connectivity
	oldLogger := commonLogger.FileLogger
	t.Cleanup(func() {
		fileLockNewFn = oldNewPid
		fileLockLockFn = oldLock
		fileLockUnlockFn = oldUnlock
		commonLoggerInitFn = oldLoggerInit
		commonLoggerCloseFn = oldLoggerClose
		loggerOpenMainFileLogFn = oldOpenLog
		isAdminFn = oldIsAdmin
		initConfigFn = oldInitCfg
		values = oldValues
		internal.Connectivity = oldConnectivity
		commonLogger.FileLogger = oldLogger
	})
	commonLogger.FileLogger = nil
	values = &server.Values{
		LogRootValues: &commonCmd.LogRootValues{},
	}
	commonLoggerInitFn = func(io.Writer) {}
	commonLoggerCloseFn = func() {}
	loggerOpenMainFileLogFn = func(string, bool) error { return nil }
	fileLockNewFn = func() *fileLock.PidLock { return &fileLock.PidLock{} }
	fileLockLockFn = func(*fileLock.PidLock) error { return nil }
	fileLockUnlockFn = func(*fileLock.PidLock) error { return nil }
	isAdminFn = func() bool { return true }
	initConfigFn = func(*pflag.FlagSet) (*internal.Configuration, string) {
		return &internal.Configuration{
			Log:            false,
			Internet:       true,
			Authentication: "disabled",
			Games: internal.Games{
				Enabled: []string{"age1"},
				Age1: internal.Game{Hosts: []string{"127.0.0.1"}},
			},
			Announcement: internal.Announcement{Enabled: false},
		}, ""
	}
}

func TestRunRootPidLockFailure(t *testing.T) {
	resetState(t)
	fileLockLockFn = func(*fileLock.PidLock) error { return errors.New("lock fail") }
	_, code := runRoot(nil)
	if code != common.ErrPidLock {
		t.Fatalf("code=%d want ErrPidLock", code)
	}
}

func TestRunRootInvalidAuth(t *testing.T) {
	resetState(t)
	initConfigFn = func(*pflag.FlagSet) (*internal.Configuration, string) {
		return &internal.Configuration{
			Log: false, Internet: true, Authentication: "invalid",
			Games: internal.Games{Enabled: []string{"age1"}, Age1: internal.Game{Hosts: []string{"127.0.0.1"}}},
			Announcement: internal.Announcement{Enabled: false},
		}, ""
	}
	_, code := runRoot(nil)
	if code != internal.ErrInvalidAuthentication {
		t.Fatalf("code=%d", code)
	}
}

func TestRunRootOpenLogFailure(t *testing.T) {
	resetState(t)
	loggerOpenMainFileLogFn = func(string, bool) error { return errors.New("log fail") }
	initConfigFn = func(*pflag.FlagSet) (*internal.Configuration, string) {
		return &internal.Configuration{
			Log: true, Internet: true, Authentication: "disabled",
			Games: internal.Games{Enabled: []string{"age1"}, Age1: internal.Game{Hosts: []string{"127.0.0.1"}}},
			Announcement: internal.Announcement{Enabled: false},
		}, ""
	}
	values.LogRoot = "/tmp"
	_, code := runRoot(nil)
	if code != common.ErrFileLog {
		t.Fatalf("code=%d", code)
	}
}

func TestRunRootNoGames(t *testing.T) {
	resetState(t)
	initConfigFn = func(*pflag.FlagSet) (*internal.Configuration, string) {
		return &internal.Configuration{
			Log: false, Internet: true, Authentication: "disabled",
			Games: internal.Games{Enabled: []string{}},
			Announcement: internal.Announcement{Enabled: false},
		}, ""
	}
	values.Id = "00000000-0000-0000-0000-000000000000"
	_, code := runRoot(nil)
	if code != internal.ErrGames {
		t.Fatalf("code=%d", code)
	}
}

func TestRunRootInvalidGame(t *testing.T) {
	resetState(t)
	initConfigFn = func(*pflag.FlagSet) (*internal.Configuration, string) {
		return &internal.Configuration{
			Log: false, Internet: true, Authentication: "disabled",
			Games: internal.Games{Enabled: []string{"invalidGame"}},
			Announcement: internal.Announcement{Enabled: false},
		}, ""
	}
	values.Id = "00000000-0000-0000-0000-000000000000"
	_, code := runRoot(nil)
	if code != internal.ErrGames {
		t.Fatalf("code=%d", code)
	}
}

func TestRunRootInvalidID(t *testing.T) {
	resetState(t)
	initConfigFn = func(*pflag.FlagSet) (*internal.Configuration, string) {
		return &internal.Configuration{
			Log: false, Internet: true, Authentication: "disabled",
			Games: internal.Games{Enabled: []string{"age1"}, Age1: internal.Game{Hosts: []string{"127.0.0.1"}}},
			Announcement: internal.Announcement{Enabled: false},
		}, ""
	}
	values.Id = "bad-uuid"
	_, code := runRoot(nil)
	if code != internal.ErrInvalidId {
		t.Fatalf("code=%d want ErrInvalidId", code)
	}
}

func TestRunRootCertFolderEmpty(t *testing.T) {
	resetState(t)
	certificatePairFolderFn = func(string) string { return "" }
	values.Id = "00000000-0000-0000-0000-000000000000"
	_, code := runRoot(nil)
	if code != internal.ErrCertDirectory {
		t.Fatalf("code=%d", code)
	}
}

func TestRunRootMulticastInvalid(t *testing.T) {
	resetState(t)
	certificatePairFolderFn = func(string) string { return t.TempDir() }
	initConfigFn = func(*pflag.FlagSet) (*internal.Configuration, string) {
		return &internal.Configuration{
			Log: false, Internet: true, Authentication: "disabled",
			Games: internal.Games{Enabled: []string{"age1"}, Age1: internal.Game{Hosts: []string{"127.0.0.1"}}},
			Announcement: internal.Announcement{Enabled: true, Multicast: true, MulticastGroup: "999.999.999.999", Port: 8080},
		}, ""
	}
	values.Id = "00000000-0000-0000-0000-000000000000"
	_, code := runRoot(nil)
	if code != internal.ErrMulticastGroup {
		t.Fatalf("code=%d want ErrMulticastGroup %d", code, internal.ErrMulticastGroup)
	}
}
