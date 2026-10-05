package cmd

import (
	"os"
	"sync"
	"testing"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/fileLock"
	"github.com/luskaner/ageLANServer/launcher-agent/internal/watch"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/agent"
)

// fakeLocker stands in for the pid lock: the test is about what the run announces,
// and a real lock would collide with an agent that happens to be running.
type fakeLocker struct{}

func (fakeLocker) Lock() error   { return nil }
func (fakeLocker) Unlock() error { return nil }

// overrides swaps every side effect runRoot reaches out for, so a test decides the
// outcome instead of waiting a minute for a process that never starts.
func overrides(t *testing.T, gameID string, exitCode int) (announced func() (string, int, int)) {
	t.Helper()
	origValues, origLock, origChdir := values, createLockFn, chdirToExeFn
	origInit, origWatch, origNotify := initializeFn, watchFn, signalNotifyFn
	origAnnounce := announce

	values = &agent.Values{
		GameIdValues:  &cmd.GameIdValues{GameId: gameID},
		LogRootValues: &cmd.LogRootValues{},
	}
	createLockFn = func() fileLock.Locker { return fakeLocker{} }
	chdirToExeFn = func() {}
	// A nil log root already skips the real initialise, but the function is
	// replaced anyway so a change to that condition cannot start writing files.
	initializeFn = func(string) {}
	signalNotifyFn = func(chan<- os.Signal, ...os.Signal) {}
	watchFn = func(*agent.Values, *watch.ExitCode, *sync.Once) {}

	var gotGame string
	var gotCode, calls int
	announce = func(game string, code int) {
		calls++
		gotGame, gotCode = game, code
	}
	// The exit code has to be produced by the watch, because that is where it comes
	// from in a real run: the caller reads it back after the cleanup.
	watchFn = func(_ *agent.Values, code *watch.ExitCode, _ *sync.Once) {
		if exitCode != common.ErrSuccess {
			code.Set(exitCode)
		}
	}

	t.Cleanup(func() {
		values, createLockFn, chdirToExeFn = origValues, origLock, origChdir
		initializeFn, watchFn, signalNotifyFn = origInit, origWatch, origNotify
		announce = origAnnounce
	})
	return func() (string, int, int) { return gotGame, gotCode, calls }
}

// A run announces its outcome, both ways.
//
// The agent outlives the launcher and prints to a console nobody is watching, so
// the notification is the only thing the user sees when the session ends. Missing
// it on the failure path would leave them with a machine they cannot account for
// and no signal that anything went wrong.
func TestRunRootAnnouncesBothOutcomes(t *testing.T) {
	for _, tc := range []struct {
		name     string
		exitCode int
	}{
		{"clean exit", common.ErrSuccess},
		{"failed exit", 42},
	} {
		t.Run(tc.name, func(t *testing.T) {
			announced := overrides(t, "age2", tc.exitCode)
			if _, exitCode := runRoot(nil); exitCode != tc.exitCode {
				t.Fatalf("exit code = %d, want %d", exitCode, tc.exitCode)
			}
			gotGame, gotCode, calls := announced()
			if calls != 1 {
				t.Fatalf("the outcome was announced %d times, want exactly 1", calls)
			}
			if gotCode != tc.exitCode {
				t.Errorf("announced exit code %d, want %d", gotCode, tc.exitCode)
			}
			if gotGame != "age2" {
				t.Errorf("announced game %q, want %q", gotGame, "age2")
			}
		})
	}
}
