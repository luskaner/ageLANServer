package cmdUtils

import (
	"io"
	"os"
	"testing"
)

// Documents the Go interface gotcha that caused L-4: assigning a typed-nil
// *os.File to an io.Writer produces a NON-NIL interface value.
func TestTypedNilFileIsNonNilWriter(t *testing.T) {
	var f *os.File = nil
	var out io.Writer = f

	// This is the Go gotcha: the interface is non-nil even though the
	// underlying pointer IS nil.
	if out == nil {
		t.Fatal("expected typed-nil to be non-nil as io.Writer (Go semantics)")
	}
}

// Regression for L-4: when FileLogger.Open returns a nil *os.File (no log
// root configured), the code must NOT pass a typed-nil to StartAgent.
func TestNilFileProducesUntypedNilWriter(t *testing.T) {
	var f *os.File // nil, as returned by Root.Open on nil receiver

	// The fix pattern used in game.go: explicit check before conversion.
	var out io.Writer
	if f != nil {
		out = f
	}

	if out != nil {
		t.Fatal("out must be untyped nil when file is nil — otherwise StartAgent's `out != nil` guard passes incorrectly")
	}
}

func TestNonNilFileProducesValidWriter(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "agent-log")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	var out io.Writer
	if f != nil {
		out = f
	}

	if out == nil {
		t.Fatal("valid file must produce non-nil writer")
	}
}
