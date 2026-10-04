package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-common/launcher"
	"github.com/spf13/pflag"
)

// recordingReporter keeps what the run told it, so a test can assert on what a
// run said instead of on where it printed it.
type recordingReporter struct {
	launcher.Discard
	steps  []string
	fails  []string
	faults []string
}

func (r *recordingReporter) Step(format string, a ...any) {
	r.steps = append(r.steps, fmt.Sprintf(format, a...))
}

func (r *recordingReporter) Fail(format string, a ...any) {
	r.fails = append(r.fails, fmt.Sprintf(format, a...))
}

func (r *recordingReporter) Fault(format string, a ...any) {
	r.faults = append(r.faults, fmt.Sprintf(format, a...))
}

// The run hands its Reporter to the shared configuration.
//
// This is the seam a graphical frontend hooks. The console installes one that
// writes to the terminal and the log file; a window installs one that writes to
// an event stream, and then reads the whole run without owning the terminal. It
// used to be impossible, because the shared operations reported to a
// package-level console logger, so the only way to see a run was to be the
// terminal it printed to.
func TestRunRootInstallsItsReporterOnTheSharedConfig(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
	})
	defer restore()

	rec := &recordingReporter{}
	origReport := report
	report = rec
	defer func() { report = origReport }()

	// Deliberately a run that goes nowhere: the exit code is not what this is
	// about. What matters is that a shared Config was handed the Reporter
	// before anything asked it to do anything.
	if _, _ = runRoot(pflag.NewFlagSet("test", pflag.ContinueOnError)); false {
		t.Fatal("unreachable")
	}
	if config.Report != launcher.Reporter(rec) {
		t.Fatalf("the shared config reports to %T, want the reporter the run installed", config.Report)
	}
}

// The console's Reporter is the same console the run has always printed to.
//
// The shared operations stopped calling the logger directly, so this is where a
// dropped line would show up: an adapter that reports into the void would keep
// every test passing and leave a user watching a terminal that says nothing.
func TestConsoleReporterStillReachesTheTerminal(t *testing.T) {
	out := captureStdout(t, func() {
		loggerReporter{}.Step("Setting up...")
		loggerReporter{}.Fail("Failed to lock pid file.")
	})
	if !strings.Contains(out, "Setting up...") {
		t.Errorf("the stage never reached the terminal, got %q", out)
	}
	if !strings.Contains(out, "Failed to lock pid file.") {
		t.Errorf("the failure never reached the terminal, got %q", out)
	}
}

// A Reporter that reports nothing must not be able to change what the run does.
// The exit code is decided by the run, not by whether anyone was listening.
func TestReporterCannotChangeTheOutcome(t *testing.T) {
	codes := make([]int, 0, 2)
	for _, r := range []launcher.Reporter{launcher.Discard{}, &recordingReporter{}} {
		restore := applyOverrides(t, runRootOverrides{
			gameId:        "age2",
			isAdmin:       false,
			gameSupported: true,
			cfg:           invalidServerStartConfig,
		})
		origReport := report
		report = r
		_, exitCode := runRoot(pflag.NewFlagSet("test", pflag.ContinueOnError))
		report = origReport
		restore()
		codes = append(codes, exitCode)
	}
	if codes[0] != codes[1] {
		t.Fatalf("the exit code depends on who was listening: %d vs %d", codes[0], codes[1])
	}
	if codes[0] == common.ErrSuccess {
		t.Fatalf("expected the run to be rejected, got success")
	}
	if codes[0] != launcher.ErrInvalidServerStart {
		t.Fatalf("exit code = %d, want %d", codes[0], launcher.ErrInvalidServerStart)
	}
}

// invalidServerStart is a configuration that cannot produce a working run, so
// the run is rejected before anything is started.
func invalidServerStartConfig() *launcher.Configuration {
	c := validLauncherConfig()
	c.Server.Start = "maybe"
	return c
}
