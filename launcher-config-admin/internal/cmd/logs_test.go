package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
)

func resetLogState(t *testing.T) {
	t.Helper()
	oldLogger := internal.Logger
	internal.Logger = nil
	t.Cleanup(func() { internal.Logger = oldLogger })
}

// Regression: runSetUp did not return after a flag parse error, so it kept
// executing the privileged setup and initialized file logging (creating the
// log directory) even for invalid invocations.
func TestRunSetUpSyntaxErrorStopsBeforeSideEffects(t *testing.T) {
	resetLogState(t)
	logRoot := filepath.Join(t.TempDir(), "logs")

	err, exitCode := runSetUp([]string{
		"--game", "age2",
		"--logRoot", logRoot,
		"--this-flag-does-not-exist",
	})

	if err == nil {
		t.Fatal("expected syntax error")
	}
	if exitCode != common.ErrSyntax {
		t.Fatalf("exitCode = %d, want ErrSyntax", exitCode)
	}
	if internal.Logger != nil {
		t.Fatal("file logging must NOT be initialized on syntax error")
	}
	if _, statErr := os.Stat(logRoot); !os.IsNotExist(statErr) {
		t.Fatal("log root directory must not be created on syntax error")
	}
}

func TestRunRevertSyntaxErrorReturnsImmediately(t *testing.T) {
	resetLogState(t)
	logRoot := filepath.Join(t.TempDir(), "logs")

	err, exitCode := runRevert([]string{
		"--logRoot", logRoot,
		"--this-flag-does-not-exist",
	})

	if err == nil {
		t.Fatal("expected syntax error")
	}
	if exitCode != common.ErrSyntax {
		t.Fatalf("exitCode = %d, want ErrSyntax", exitCode)
	}
	if internal.Logger != nil {
		t.Fatal("file logging must NOT be initialized on syntax error")
	}
	if _, statErr := os.Stat(logRoot); !os.IsNotExist(statErr) {
		t.Fatal("log root directory must not be created on syntax error")
	}
}
