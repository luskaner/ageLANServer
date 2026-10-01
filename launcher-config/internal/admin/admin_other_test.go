//go:build !windows

package admin

import (
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/luskaner/ageLANServer/common/executor"
)

func saveOtherDeps(t *testing.T) {
	t.Helper()
	origWait := processWaitForProcessFn
	origProcess := processFn
	origSleep := sleepFn
	t.Cleanup(func() {
		processWaitForProcessFn = origWait
		processFn = origProcess
		sleepFn = origSleep
	})
}

func TestPostAgentStartAdminReturnsTrue(t *testing.T) {
	if !executor.IsAdmin() {
		t.Skip("needs to look elevated to take the admin shortcut")
	}
	if !postAgentStart(123, "file") {
		t.Fatal("an already elevated process has nothing to wait for")
	}
}

func TestPostAgentStartSucceedsWhenAgentAppears(t *testing.T) {
	saveOtherDeps(t)
	processWaitForProcessFn = func(*os.Process, *time.Duration) bool { return false }
	calls := 0
	processFn = func(string) (string, *os.Process, error) {
		calls++
		if calls < 3 {
			return "", nil, nil // not up yet
		}
		return "/tmp/pid", &os.Process{Pid: 7}, nil
	}
	sleepFn = func(time.Duration) {}

	if !postAgentStart(123, "file") {
		t.Fatal("expected success once the agent pid file appears")
	}
	if calls != 3 {
		t.Errorf("polled %d times, want 3", calls)
	}
}

// A pid that is already gone means the agent died instead of starting. That has
// to be reported instead of retried until the budget runs out.
func TestPostAgentStartFailsFastOnDeadPid(t *testing.T) {
	saveOtherDeps(t)
	// WaitForProcess waits for exit, so true means it died straight away.
	processWaitForProcessFn = func(*os.Process, *time.Duration) bool { return true }
	processFn = func(string) (string, *os.Process, error) { return "", nil, nil }
	slept := 0
	sleepFn = func(time.Duration) { slept++ }

	if postAgentStart(123, "file") {
		t.Fatal("expected failure when the agent pid exited immediately")
	}
	if slept != 0 {
		t.Errorf("slept %d times, want 0: a dead pid must not consume the whole budget", slept)
	}
}

// Regression: the loop was unbounded with no failure path, so an agent that
// never came up left config.exe hanging forever with nothing printed.
func TestPostAgentStartGivesUpWhenAgentNeverAppears(t *testing.T) {
	saveOtherDeps(t)
	if postAgentStartAttempts <= 0 {
		t.Fatal("postAgentStartAttempts must be positive for the loop to progress")
	}
	processWaitForProcessFn = func(*os.Process, *time.Duration) bool { return false }
	processFn = func(string) (string, *os.Process, error) { return "", nil, nil }
	polls := 0
	sleepFn = func(time.Duration) { polls++ }

	if postAgentStart(123, "file") {
		t.Fatal("expected failure when the agent never appears")
	}
	if polls != postAgentStartAttempts {
		t.Errorf("polled %d times, want the whole budget of %d", polls, postAgentStartAttempts)
	}
}

func TestIsAccessDeniedOther(t *testing.T) {
	if !isAccessDenied(os.ErrPermission) {
		t.Error("os.ErrPermission should be reported as an access denial")
	}
	if isAccessDenied(nil) {
		t.Error("nil is not an access denial")
	}
	if isAccessDenied(net.ErrClosed) {
		t.Error("net.ErrClosed is not an access denial")
	}
	if isAccessDenied(errors.New("something else")) {
		t.Error("an unrelated error is not an access denial")
	}
}
