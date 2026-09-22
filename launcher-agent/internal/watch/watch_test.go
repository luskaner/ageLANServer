package watch

import (
	"io"
	"os"
	gos_exec "os/exec"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher-agent/internal"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/agent"
)

// Helper process: re-executed by the tests to create real child processes
// that sleep for SLEEP_MS milliseconds and exit successfully.
func TestHelperProcessSleep(t *testing.T) {
	if os.Getenv("GO_WATCH_HELPER") != "1" {
		return
	}
	ms, _ := strconv.Atoi(os.Getenv("SLEEP_MS"))
	time.Sleep(time.Duration(ms) * time.Millisecond)
	os.Exit(0)
}

func spawnSleeper(t *testing.T, ms int) *os.Process {
	t.Helper()
	cmd := gos_exec.Command(os.Args[0], "-test.run=TestHelperProcessSleep$", "-test.v")
	cmd.Env = append(os.Environ(), "GO_WATCH_HELPER=1", "SLEEP_MS="+strconv.Itoa(ms))
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })
	return cmd.Process
}

func TestWaitForProcessesToExitWaitsForAll(t *testing.T) {
	fast := spawnSleeper(t, 50)
	slow := spawnSleeper(t, 500)

	start := time.Now()
	if !waitForProcessesToExit([]*os.Process{fast, slow}) {
		t.Fatal("expected successful wait")
	}
	elapsed := time.Since(start)
	if elapsed < 400*time.Millisecond {
		t.Fatalf("returned after %v; it did not wait for ALL processes (regression: single random pick)", elapsed)
	}
}

func TestWaitForProcessesToExitEmptyAndNil(t *testing.T) {
	if !waitForProcessesToExit(nil) {
		t.Fatal("empty list must succeed")
	}
	if !waitForProcessesToExit([]*os.Process{nil}) {
		t.Fatal("nil entries must be skipped")
	}
}

func TestSortedProcessNames(t *testing.T) {
	processes := map[string]*os.Process{
		"zebra.exe": nil,
		"alpha.exe": nil,
		"mid.exe":   nil,
	}
	names := sortedProcessNames(processes)
	want := []string{"alpha.exe", "mid.exe", "zebra.exe"}
	if len(names) != len(want) {
		t.Fatalf("names = %v", names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("names = %v, want %v", names, want)
		}
	}
}

func testValues() *agent.Values {
	return &agent.Values{
		GameIdValues:  &cmd.GameIdValues{GameId: "age2"},
		LogRootValues: &cmd.LogRootValues{LogRoot: ""},
	}
}

type watchOverrides struct {
	waitUntilAnyProcessExistFnVal func([]string) map[string]*os.Process
	waitForProcessesToExitFnVal   func([]*os.Process) bool
	serverKillDoFnVal             func(string) error
	configRevertFnVal             func(string, string, bool, io.Writer, func(*exec.Options), func([]string, bool, io.Writer, func(*exec.Options)) *exec.Result) bool
	runRevertCommandFnVal         func(io.Writer, func(*exec.Options)) error
	removeBattleServerRegionFnVal func(string, string, string, io.Writer, func(*exec.Options)) *exec.Result
	loggerBufferFnVal             func(string, func(io.Writer)) error
	gameLogsCopyFnVal             func(string, string, string)
	rebroadcastFnVal              func(*int, int)
}

func applyWatchOverrides(o watchOverrides) func() {
	origProcessFn := waitUntilAnyProcessExistFn
	origWaitFn := waitForProcessesToExitFn
	origKillFn := serverKillDoFn
	origConfigRevert := configRevertFn
	origRevertCommand := runRevertCommandFn
	origRemoveBSRegion := removeBattleServerRegionFn
	origBufferFn := loggerBufferFn
	origCopyLogs := gameLogsCopyFn
	origRebroadcast := rebroadcastFn

	if o.waitUntilAnyProcessExistFnVal != nil {
		waitUntilAnyProcessExistFn = o.waitUntilAnyProcessExistFnVal
	}
	if o.waitForProcessesToExitFnVal != nil {
		waitForProcessesToExitFn = o.waitForProcessesToExitFnVal
	}
	if o.serverKillDoFnVal != nil {
		serverKillDoFn = o.serverKillDoFnVal
	}
	if o.configRevertFnVal != nil {
		configRevertFn = o.configRevertFnVal
	}
	if o.runRevertCommandFnVal != nil {
		runRevertCommandFn = o.runRevertCommandFnVal
	}
	if o.removeBattleServerRegionFnVal != nil {
		removeBattleServerRegionFn = o.removeBattleServerRegionFnVal
	}
	if o.loggerBufferFnVal != nil {
		loggerBufferFn = o.loggerBufferFnVal
	}
	if o.gameLogsCopyFnVal != nil {
		gameLogsCopyFn = o.gameLogsCopyFnVal
	}
	if o.rebroadcastFnVal != nil {
		rebroadcastFn = o.rebroadcastFnVal
	}

	return func() {
		waitUntilAnyProcessExistFn = origProcessFn
		waitForProcessesToExitFn = origWaitFn
		serverKillDoFn = origKillFn
		configRevertFn = origConfigRevert
		runRevertCommandFn = origRevertCommand
		removeBattleServerRegionFn = origRemoveBSRegion
		loggerBufferFn = origBufferFn
		gameLogsCopyFn = origCopyLogs
		rebroadcastFn = origRebroadcast
	}
}

func noopConfigRevert(gameId, logRoot string, headless bool, out io.Writer, optionsFn func(*exec.Options), runRevertFn func([]string, bool, io.Writer, func(*exec.Options)) *exec.Result) bool {
	return true
}

func noopRevertCommand(out io.Writer, optionsFn func(*exec.Options)) error {
	return nil
}

func noopLoggerBuffer(name string, fn func(io.Writer)) error {
	return nil
}

func noopRemoveBSRegion(exe, gameId, region string, out io.Writer, optionsFn func(*exec.Options)) *exec.Result {
	return &exec.Result{}
}

func TestWatchGameTimeoutStart(t *testing.T) {
	defer applyWatchOverrides(watchOverrides{
		waitUntilAnyProcessExistFnVal: func(names []string) map[string]*os.Process {
			return nil
		},
		serverKillDoFnVal:     func(name string) error { return nil },
		configRevertFnVal:     noopConfigRevert,
		runRevertCommandFnVal: noopRevertCommand,
		loggerBufferFnVal:     noopLoggerBuffer,
		gameLogsCopyFnVal:     func(gameId, basePath, logRoot string) {},
	})()

	values := testValues()
	values.ServerExecutable = ""
	exitCode := 0
	var once sync.Once
	Watch(values, &exitCode, &once)
	if exitCode != internal.ErrGameTimeoutStart {
		t.Errorf("expected %d for game timeout, got %d", internal.ErrGameTimeoutStart, exitCode)
	}
}

func TestWatchGameFound(t *testing.T) {
	fakeProcess := &os.Process{Pid: 99999}
	defer applyWatchOverrides(watchOverrides{
		waitUntilAnyProcessExistFnVal: func(names []string) map[string]*os.Process {
			return map[string]*os.Process{"game.exe": fakeProcess}
		},
		waitForProcessesToExitFnVal: func(processes []*os.Process) bool { return true },
		serverKillDoFnVal:           func(name string) error { return nil },
		configRevertFnVal:           noopConfigRevert,
		runRevertCommandFnVal:       noopRevertCommand,
		loggerBufferFnVal:           noopLoggerBuffer,
		gameLogsCopyFnVal:           func(gameId, basePath, logRoot string) {},
		rebroadcastFnVal:            func(exitCode *int, port int) {},
	})()

	values := testValues()
	values.ServerExecutable = ""
	exitCode := 0
	var once sync.Once
	Watch(values, &exitCode, &once)
	if exitCode != common.ErrSuccess {
		t.Errorf("expected success, got %d", exitCode)
	}
}

func TestWatchWithServerKill(t *testing.T) {
	fakeProcess := &os.Process{Pid: 99999}
	killCalled := false
	defer applyWatchOverrides(watchOverrides{
		waitUntilAnyProcessExistFnVal: func(names []string) map[string]*os.Process {
			return map[string]*os.Process{"game.exe": fakeProcess}
		},
		waitForProcessesToExitFnVal: func(processes []*os.Process) bool { return true },
		serverKillDoFnVal: func(name string) error {
			killCalled = true
			return nil
		},
		configRevertFnVal:     noopConfigRevert,
		runRevertCommandFnVal: noopRevertCommand,
		loggerBufferFnVal:     noopLoggerBuffer,
		gameLogsCopyFnVal:     func(gameId, basePath, logRoot string) {},
	})()

	values := testValues()
	values.ServerExecutable = "server.exe"
	exitCode := 0
	var once sync.Once
	Watch(values, &exitCode, &once)
	if !killCalled {
		t.Error("expected serverKill to be called")
	}
	if exitCode != common.ErrSuccess {
		t.Errorf("expected success, got %d", exitCode)
	}
}

func TestWatchServerKillFailure(t *testing.T) {
	fakeProcess := &os.Process{Pid: 99999}
	defer applyWatchOverrides(watchOverrides{
		waitUntilAnyProcessExistFnVal: func(names []string) map[string]*os.Process {
			return map[string]*os.Process{"game.exe": fakeProcess}
		},
		waitForProcessesToExitFnVal: func(processes []*os.Process) bool { return true },
		serverKillDoFnVal: func(name string) error {
			return os.ErrProcessDone
		},
		configRevertFnVal:     noopConfigRevert,
		runRevertCommandFnVal: noopRevertCommand,
		loggerBufferFnVal:     noopLoggerBuffer,
		gameLogsCopyFnVal:     func(gameId, basePath, logRoot string) {},
	})()

	values := testValues()
	values.ServerExecutable = "server.exe"
	exitCode := 0
	var once sync.Once
	Watch(values, &exitCode, &once)
	if exitCode != internal.ErrFailedStopServer {
		t.Errorf("expected %d for kill failure, got %d", internal.ErrFailedStopServer, exitCode)
	}
}

func TestWatchWithLogCopy(t *testing.T) {
	fakeProcess := &os.Process{Pid: 99999}
	logCopied := false
	defer applyWatchOverrides(watchOverrides{
		waitUntilAnyProcessExistFnVal: func(names []string) map[string]*os.Process {
			return map[string]*os.Process{"game.exe": fakeProcess}
		},
		waitForProcessesToExitFnVal: func(processes []*os.Process) bool { return true },
		serverKillDoFnVal:           func(name string) error { return nil },
		configRevertFnVal:           noopConfigRevert,
		runRevertCommandFnVal:       noopRevertCommand,
		loggerBufferFnVal:           noopLoggerBuffer,
		gameLogsCopyFnVal: func(gameId, basePath, logRoot string) {
			logCopied = true
		},
	})()

	values := testValues()
	values.ServerExecutable = ""
	values.LogRoot = "/tmp/logs"
	values.BaseDataPath = "/tmp/data"
	exitCode := 0
	var once sync.Once
	Watch(values, &exitCode, &once)
	if !logCopied {
		t.Error("expected game logs to be copied")
	}
}

func TestWatchCleanupOncePreventsDoubleCleanup(t *testing.T) {
	fakeProcess := &os.Process{Pid: 99999}
	killCount := 0
	defer applyWatchOverrides(watchOverrides{
		waitUntilAnyProcessExistFnVal: func(names []string) map[string]*os.Process {
			return map[string]*os.Process{"game.exe": fakeProcess}
		},
		waitForProcessesToExitFnVal: func(processes []*os.Process) bool { return true },
		serverKillDoFnVal: func(name string) error {
			killCount++
			return nil
		},
		configRevertFnVal:     noopConfigRevert,
		runRevertCommandFnVal: noopRevertCommand,
		loggerBufferFnVal:     noopLoggerBuffer,
		gameLogsCopyFnVal:     func(gameId, basePath, logRoot string) {},
	})()

	values := testValues()
	values.ServerExecutable = "server.exe"
	exitCode := 0
	var once sync.Once
	Watch(values, &exitCode, &once)
	if killCount != 1 {
		t.Errorf("expected serverKill to be called exactly once, got %d", killCount)
	}
}

func TestWatchWaitFailure(t *testing.T) {
	fakeProcess := &os.Process{Pid: 99999}
	defer applyWatchOverrides(watchOverrides{
		waitUntilAnyProcessExistFnVal: func(names []string) map[string]*os.Process {
			return map[string]*os.Process{"game.exe": fakeProcess}
		},
		waitForProcessesToExitFnVal: func(processes []*os.Process) bool { return false },
		serverKillDoFnVal:           func(name string) error { return nil },
		configRevertFnVal:           noopConfigRevert,
		runRevertCommandFnVal:       noopRevertCommand,
		loggerBufferFnVal:           noopLoggerBuffer,
		gameLogsCopyFnVal:           func(gameId, basePath, logRoot string) {},
	})()

	values := testValues()
	values.ServerExecutable = ""
	exitCode := 0
	var once sync.Once
	Watch(values, &exitCode, &once)
	if exitCode != internal.ErrFailedWaitForProcess {
		t.Errorf("expected %d for wait failure, got %d", internal.ErrFailedWaitForProcess, exitCode)
	}
}

func TestWaitUntilAnyProcessExistFound(t *testing.T) {
	origFn := commonProcessProcessesByNamesFn
	defer func() { commonProcessProcessesByNamesFn = origFn }()

	fakeProcess := &os.Process{Pid: 12345}
	commonProcessProcessesByNamesFn = func(names []string) map[string]*os.Process {
		return map[string]*os.Process{"game.exe": fakeProcess}
	}

	processes := waitUntilAnyProcessExist([]string{"game.exe"})
	if len(processes) != 1 {
		t.Fatalf("expected 1 process, got %d", len(processes))
	}
	if _, ok := processes["game.exe"]; !ok {
		t.Fatal("expected game.exe in results")
	}
}

func TestWaitUntilAnyProcessExistTimeout(t *testing.T) {
	origFn := commonProcessProcessesByNamesFn
	origInterval := processWaitInterval
	defer func() {
		commonProcessProcessesByNamesFn = origFn
		processWaitInterval = origInterval
	}()

	processWaitInterval = 50 * time.Millisecond
	commonProcessProcessesByNamesFn = func(names []string) map[string]*os.Process {
		return nil
	}

	origOneMinute := oneMinuteWaitTimeout
	oneMinuteWaitTimeout = 200 * time.Millisecond
	defer func() { oneMinuteWaitTimeout = origOneMinute }()

	start := time.Now()
	processes := waitUntilAnyProcessExist([]string{"nonexistent.exe"})
	elapsed := time.Since(start)
	if len(processes) != 0 {
		t.Fatalf("expected empty result, got %d processes", len(processes))
	}
	if elapsed < 150*time.Millisecond {
		t.Fatalf("expected at least 150ms wait, got %v", elapsed)
	}
}
