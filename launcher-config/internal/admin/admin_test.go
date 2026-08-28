package admin

import (
	"crypto/x509"
	"errors"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-config/internal"
)

// resetAdminState cleans global ipc state and restores DI vars.
func resetAdminState(t *testing.T) {
	t.Helper()
	oldIPC := ipc
	oldEnc := encoder
	oldDec := decoder
	oldBytes := bytesToCertificateFn
	oldNewFile := newFileFn
	oldRunSetUp := runSetUpFn
	oldRunRevert := runRevertFn
	oldRunFlush := runFlushCacheFn
	oldRunFlushAgent := runFlushCacheAgentFn
	oldProcess := processFn
	oldKill := killPidProcFn
	oldDial := dialIPCFn
	oldPost := postAgentStartFn
	oldNative := nativeFileNameFn
	oldSleep := sleepFn
	oldConnect := connectAgentIfNeededFn
	oldGetFolder := getLoggerFolderFn
	oldLogger := internal.Logger
	t.Cleanup(func() {
		ipc = oldIPC
		encoder = oldEnc
		decoder = oldDec
		bytesToCertificateFn = oldBytes
		newFileFn = oldNewFile
		runSetUpFn = oldRunSetUp
		runRevertFn = oldRunRevert
		runFlushCacheFn = oldRunFlush
		runFlushCacheAgentFn = oldRunFlushAgent
		processFn = oldProcess
		killPidProcFn = oldKill
		dialIPCFn = oldDial
		postAgentStartFn = oldPost
		nativeFileNameFn = oldNative
		sleepFn = oldSleep
		connectAgentIfNeededFn = oldConnect
		getLoggerFolderFn = oldGetFolder
		internal.Logger = oldLogger
	})
	// reset to known clean state
	ipc = nil
	encoder = nil
	decoder = nil
	internal.Logger = nil
	sleepFn = func(time.Duration) {}
}

func TestRunSetUpCertParseFailure(t *testing.T) {
	resetAdminState(t)
	bytesToCertificateFn = func([]byte) *x509.Certificate { return nil }
	_, exitCode := RunSetUp("age2", "", net.ParseIP("127.0.0.1"), false, []byte("bad"))
	if exitCode != internal.ErrUserCertAddParse {
		t.Fatalf("exitCode = %d, want %d", exitCode, internal.ErrUserCertAddParse)
	}
}

func TestRunSetUpNewFileFailure(t *testing.T) {
	resetAdminState(t)
	newFileFn = func(string, string, bool) (error, *commonLogger.Root) {
		return errors.New("new file fail"), nil
	}
	_, exitCode := RunSetUp("age2", "/tmp/log", nil, false, nil)
	if exitCode != common.ErrFileLog {
		t.Fatalf("exitCode = %d, want ErrFileLog %d", exitCode, common.ErrFileLog)
	}
}

func TestRunSetUpSuccessNoLog(t *testing.T) {
	resetAdminState(t)
	runSetUpFn = func(string, net.IP, bool, *x509.Certificate, string, io.Writer, func(*exec.Options)) *exec.Result {
		return &exec.Result{ExitCode: common.ErrSuccess}
	}
	_, exitCode := RunSetUp("age2", "", net.ParseIP("127.0.0.1"), false, nil)
	if exitCode != common.ErrSuccess {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}
}

func TestRunSetUpSuccessWithLog(t *testing.T) {
	resetAdminState(t)
	tmp := t.TempDir()
	newFileFn = func(root, gameId string, finalRoot bool) (error, *commonLogger.Root) {
		return commonLogger.NewFile(tmp, "", true)
	}
	runSetUpFn = func(string, net.IP, bool, *x509.Certificate, string, io.Writer, func(*exec.Options)) *exec.Result {
		return &exec.Result{ExitCode: common.ErrSuccess}
	}
	_, exitCode := RunSetUp("age2", tmp, net.ParseIP("127.0.0.1"), false, nil)
	if exitCode != common.ErrSuccess {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}
}

func TestRunRevertNewFileFailure(t *testing.T) {
	resetAdminState(t)
	newFileFn = func(string, string, bool) (error, *commonLogger.Root) {
		return errors.New("fail"), nil
	}
	_, exitCode := RunRevert("/tmp/log", true, true, true)
	if exitCode != common.ErrFileLog {
		t.Fatalf("exitCode = %d, want ErrFileLog", exitCode)
	}
}

func TestRunRevertSuccessNoLog(t *testing.T) {
	resetAdminState(t)
	runRevertFn = func(bool, bool, bool, string, io.Writer, func(*exec.Options)) *exec.Result {
		return &exec.Result{ExitCode: common.ErrSuccess}
	}
	_, exitCode := RunRevert("", true, false, true)
	if exitCode != common.ErrSuccess {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}
}

func TestRunFlushCacheAgentAlreadyStarted(t *testing.T) {
	resetAdminState(t)
	// Simulate agent already connected by setting ipc non-nil
	ipc = &mockConn{}
	_, exitCode := RunFlushCache("", true, true)
	if exitCode != internal.ErrAgentAlreadyStarted {
		t.Fatalf("exitCode = %d, want ErrAgentAlreadyStarted %d", exitCode, internal.ErrAgentAlreadyStarted)
	}
}

func TestRunFlushCacheNewFileFailure(t *testing.T) {
	resetAdminState(t)
	newFileFn = func(string, string, bool) (error, *commonLogger.Root) {
		return errors.New("fail"), nil
	}
	_, exitCode := RunFlushCache("/tmp/log", true, true)
	if exitCode != common.ErrFileLog {
		t.Fatalf("exitCode = %d, want ErrFileLog", exitCode)
	}
}

func TestRunFlushCacheSuccessNoLog(t *testing.T) {
	resetAdminState(t)
	runFlushCacheFn = func(bool, bool, string, io.Writer, func(*exec.Options)) (string, *exec.Result) {
		return "", &exec.Result{ExitCode: common.ErrSuccess}
	}
	_, exitCode := RunFlushCache("", true, true)
	if exitCode != common.ErrSuccess {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}
}

func TestConnectAgentAlreadyConnected(t *testing.T) {
	resetAdminState(t)
	ipc = &mockConn{}
	if err := ConnectAgentIfNeeded(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestConnectAgentDialFailure(t *testing.T) {
	resetAdminState(t)
	dialIPCFn = func() (net.Conn, error) { return nil, errors.New("dial fail") }
	if err := ConnectAgentIfNeeded(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnectAgentDialSuccess(t *testing.T) {
	resetAdminState(t)
	// Use net.Pipe to get a real conn that can be closed
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()
	dialIPCFn = func() (net.Conn, error) { return c1, nil }
	if err := ConnectAgentIfNeeded(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if ipc != c1 {
		t.Fatal("ipc not set")
	}
	if encoder == nil || decoder == nil {
		t.Fatal("encoder/decoder not set")
	}
	// cleanup will close
}

func TestStopAgentWhenNotConnectedAndNoProcess(t *testing.T) {
	resetAdminState(t)
	connectAgentIfNeededFn = func() error { return errors.New("not connected") }
	nativeFileNameFn = func(bool, string) string { return "dummy.exe" }
	processFn = func(string) (string, *os.Process, error) { return "", nil, nil }
	if !StopAgentIfNeeded() {
		t.Fatal("expected true when no process")
	}
}

func TestStopAgentKillSuccess(t *testing.T) {
	resetAdminState(t)
	connectAgentIfNeededFn = func() error { return errors.New("not connected") }
	nativeFileNameFn = func(bool, string) string { return "dummy.exe" }
	// First call (check not connected) returns no proc => not taken, then stopAgentIfNeeded fails? Let's test kill path:
	// StopAgentIfNeeded flow: if not connected, check process -> if proc nil return true. So to reach kill, need proc not nil.
	// Instead test kill path after stopAgentIfNeeded fails.
	// We can force connectAgent to return nil (connected) so first early return not taken.
	connectAgentIfNeededFn = func() error { return nil }
	// stopAgentIfNeeded will be called; it will see ipc != nil? But we set ipc nil, so it returns nil without encoding.
	// Then loop 30 checks process -> we make it return nil first 30 to simulate not stopped, then kill.
	called := 0
	processFn = func(string) (string, *os.Process, error) {
		called++
		// First 31 calls (inside loop) return nil proc to simulate not stopped.
		// After loop, final check expects proc not nil to kill.
		if called <= 31 {
			return "", nil, nil
		}
		return "/tmp/pid", &os.Process{Pid: 999}, nil
	}
	killPidProcFn = func(string, *os.Process) error { return nil }
	// Need to set ipc to mock to make stopAgentIfNeeded encode path? Actually ipc nil will make stopAgentIfNeeded do nothing and return nil, so it will enter loop.
	// The loop will check process 30 times, all nil, then fail, then final process returns proc, kill succeeds.
	// But our called count will be 31+? Let's make first 30 nil, then final kill true.
	processFn = func(string) (string, *os.Process, error) {
		if called < 30 {
			called++
			return "", nil, nil
		}
		if called == 30 {
			called++
			return "/tmp/pid", &os.Process{Pid: 999}, nil
		}
		called++
		return "", nil, nil
	}
	// Actually simpler: make loop succeed? Let's test the kill path directly:
	// Make stopAgentIfNeeded return error to skip loop, then go to kill
	// We can set ipc to non-nil and encoder to mock that fails? That's complicated.
	// For now just test the simple path where process is found and killed.
	resetAdminState(t)
	connectAgentIfNeededFn = func() error { return errors.New("not") }
	nativeFileNameFn = func(bool, string) string { return "dummy.exe" }
	processFn = func(string) (string, *os.Process, error) {
		return "/tmp/pid", &os.Process{Pid: 1}, nil
	}
	killPidProcFn = func(string, *os.Process) error { return nil }
	// Need to ensure first check (not connected) doesn't return true: it checks process returns nil => we return proc not nil, so first check fails
	// Then it tries stopAgentIfNeeded: with ipc nil, it returns nil, then loop 30: process returns proc not nil? Actually loop checks for proc == nil to return true.
	// If process always returns proc, loop never returns, then after loop it checks and kills.
	// Let's make loop always find proc, so it never returns early, then final kill succeeds.
	processFn = func(string) (string, *os.Process, error) {
		return "/tmp/pid", &os.Process{Pid: 1}, nil
	}
	// sleep no-op
	if !StopAgentIfNeeded() {
		t.Fatal("expected true after kill")
	}
}

func TestStartAgentSuccess(t *testing.T) {
	resetAdminState(t)
	getLoggerFolderFn = func() string { return "" }
	runFlushCacheAgentFn = func(bool, bool, string, io.Writer, func(*exec.Options)) (string, *exec.Result) {
		return "/tmp/file", &exec.Result{ExitCode: common.ErrSuccess, Pid: 123}
	}
	postAgentStartFn = func(uint32, string) bool { return true }
	result := StartAgent(true, true)
	if !result.Success() {
		t.Fatalf("expected success, got %+v", result)
	}
}

func TestStartAgentPostStartFailure(t *testing.T) {
	resetAdminState(t)
	getLoggerFolderFn = func() string { return "" }
	runFlushCacheAgentFn = func(bool, bool, string, io.Writer, func(*exec.Options)) (string, *exec.Result) {
		return "/tmp/file", &exec.Result{ExitCode: common.ErrSuccess, Pid: 123}
	}
	postAgentStartFn = func(uint32, string) bool { return false }
	result := StartAgent(true, true)
	if result.Err == nil {
		t.Fatal("expected error when postAgentStart fails")
	}
}

func TestStartAgentFailure(t *testing.T) {
	resetAdminState(t)
	getLoggerFolderFn = func() string { return "" }
	runFlushCacheAgentFn = func(bool, bool, string, io.Writer, func(*exec.Options)) (string, *exec.Result) {
		return "", &exec.Result{Err: errors.New("start fail"), ExitCode: 1}
	}
	result := StartAgent(true, true)
	if result.Success() {
		t.Fatal("expected failure")
	}
}

func TestConnectWithRetriesSuccess(t *testing.T) {
	resetAdminState(t)
	connectAgentIfNeededFn = func() error { return nil }
	if !ConnectAgentIfNeededWithRetries() {
		t.Fatal("expected true")
	}
}

func TestConnectWithRetriesFailure(t *testing.T) {
	resetAdminState(t)
	connectAgentIfNeededFn = func() error { return errors.New("fail") }
	if ConnectAgentIfNeededWithRetries() {
		t.Fatal("expected false")
	}
}

func TestClearIPCState(t *testing.T) {
	resetAdminState(t)
	c1, c2 := net.Pipe()
	defer c2.Close()
	ipc = c1
	encoder = nil
	decoder = nil
	clearIPCState()
	if ipc != nil {
		t.Fatal("ipc should be nil")
	}
}

// mockConn implements net.Conn for tests
type mockConn struct{ net.Conn }

func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// need x509 import
//go:generate
