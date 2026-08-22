package launcher_common

import (
	"os"
	"testing"

	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func storeCommand(t *testing.T, cmd []string) {
	t.Helper()
	if err := RevertCommandStore.Delete(); err != nil {
		t.Fatal(err)
	}
	if err := RevertCommandStore.Store(cmd); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = RevertCommandStore.Delete() })
}

// Regression: RunRevertCommand deleted the stored command even when execution
// FAILED, losing the elevated action forever. It must only clear the store on
// success, mirroring ConfigRevert.
func TestRunRevertCommandKeepsStoreOnFailure(t *testing.T) {
	storeCommand(t, []string{"fake-exe", "--do-stuff"})

	var received exec.Options
	oldExec := revertCommandExec
	revertCommandExec = func(options exec.Options) *exec.Result {
		received = options
		return &exec.Result{Err: os.ErrPermission, ExitCode: 1}
	}
	defer func() { revertCommandExec = oldExec }()

	err := RunRevertCommand(nil, nil)
	if err == nil {
		t.Fatal("expected the failure error to propagate")
	}
	if received.File != "fake-exe" || len(received.Args) != 1 || received.Args[0] != "--do-stuff" {
		t.Fatalf("options not forwarded: %+v", received)
	}

	err, flags := RevertCommandStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(flags) == 0 {
		t.Fatal("store was deleted on failure; failed command is lost and can never be retried")
	}
}

func TestRunRevertCommandClearsStoreOnSuccess(t *testing.T) {
	storeCommand(t, []string{"fake-exe", "--ok"})

	oldExec := revertCommandExec
	revertCommandExec = func(exec.Options) *exec.Result {
		return &exec.Result{}
	}
	defer func() { revertCommandExec = oldExec }()

	if err := RunRevertCommand(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err, flags := RevertCommandStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(flags) != 0 {
		t.Fatalf("store must be cleared on success, still has %v", flags)
	}
}

func TestRunRevertCommandNoopWithoutStore(t *testing.T) {
	if err := RevertCommandStore.Delete(); err != nil {
		t.Fatal(err)
	}
	execRan := false
	oldExec := revertCommandExec
	revertCommandExec = func(exec.Options) *exec.Result {
		execRan = true
		return &exec.Result{}
	}
	defer func() { revertCommandExec = oldExec }()

	if err := RunRevertCommand(nil, nil); err != nil {
		t.Fatal(err)
	}
	if execRan {
		t.Fatal("must not execute anything without a stored command")
	}
}
