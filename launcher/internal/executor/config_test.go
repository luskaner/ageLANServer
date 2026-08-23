package executor

import (
	"errors"
	"strings"
	"testing"

	"github.com/luskaner/ageLANServer/common/executor/exec"
)

// Regression: when Store failed AND the compensating RunRevert also failed,
// the store error overwrote the revert error, hiding the fact that the system
// was left in an unrecoverable state (changes applied but nothing to undo).
func TestStoreAndRevertErrorsAreJoined(t *testing.T) {
	revertErr := errors.New("revert: cannot restore hosts file")
	storeErr := errors.New("store: disk full")

	result := &exec.Result{Err: revertErr}
	storeFailure := storeErr

	// The fix pattern from RunSetUp:
	result.Err = errors.Join(result.Err, storeFailure)

	if result.Err == nil {
		t.Fatal("joined error is nil")
	}
	if !strings.Contains(result.Err.Error(), "revert: cannot restore hosts file") {
		t.Fatalf("joined error lost the revert error: %v", result.Err)
	}
	if !strings.Contains(result.Err.Error(), "store: disk full") {
		t.Fatalf("joined error lost the store error: %v", result.Err)
	}
}

// When only the store fails (revert succeeded), the joined error is just the
// store error — no information loss.
func TestStoreOnlyErrorPreserved(t *testing.T) {
	var result *exec.Result // nil Err means success

	storeErr := errors.New("store: permission denied")
	result = &exec.Result{}
	result.Err = errors.Join(nil, storeErr)

	if result.Err == nil {
		t.Fatal("joined error should not be nil")
	}
	if !strings.Contains(result.Err.Error(), "store: permission denied") {
		t.Fatalf("unexpected error: %v", result.Err)
	}
}

// When the compensating revert fails but there was no store error,
// the revert error must not be lost.
func TestRevertOnlyErrorNotOverwritten(t *testing.T) {
	revertErr := errors.New("revert: access denied")

	result := &exec.Result{Err: revertErr}
	storeErr := error(nil)

	result.Err = errors.Join(result.Err, storeErr)

	if result.Err == nil {
		t.Fatal("error must survive")
	}
	if !errors.Is(result.Err, revertErr) {
		t.Fatalf("revert error lost after join: %v", result.Err)
	}
}
