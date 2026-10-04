// Package dialog holds the launcher console's answers to the interactive
// prompts. The prompts themselves, and the Dialog interface, live in
// launcher-common/launcher: every frontend asks the same questions, and only the
// way they are asked differs.
//
// It must not import cmdUtils or cmdUtils/logger: cmdUtils needs dialog, and
// cmdUtils/logger needs dialog too, so console output goes through the
// injectable Output sinks instead.
package dialog

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/luskaner/ageLANServer/launcher-common/launcher"
)

// The vocabulary belongs to the shared package, so a graphical frontend answers
// the same questions with the same types. These aliases keep this package's
// callers saying dialog.ServerCandidate instead of repeating where it came from.
type (
	Dialog          = launcher.Dialog
	ServerCandidate = launcher.ServerCandidate
)

const (
	ModeAuto  = launcher.ModeAuto
	ModeTrue  = launcher.ModeTrue
	ModeFalse = launcher.ModeFalse
)

// label is the compact summary of a candidate, falling back to the full
// description when no compact one was provided.
//
// It is a backend detail rather than a method on the candidate: which of the two
// forms fits depends on how much room the backend has.
func label(c ServerCandidate) string {
	if c.Label == "" {
		return c.Description
	}
	return c.Label
}

// Output are the sinks the console prompts write to. They mirror
// cmdUtils/logger so prompt output keeps reaching both the log file and stdout.
type Output struct {
	Println func(...any)
	Printf  func(string, ...any)
}

// Resolution is the outcome of resolving the configured mode to a backend.
type Resolution struct {
	Dialog Dialog
	// Name is "zenity" or "console".
	Name string
	// Reason is non-empty when the configured mode could not be honoured and
	// the console fallback was used instead. It is meant to be logged.
	Reason string
}

var (
	active atomic.Pointer[Dialog]

	outputMu sync.RWMutex
	output   = defaultOutput()
)

func defaultOutput() Output {
	return Output{
		Println: func(a ...any) { fmt.Println(a...) },
		Printf:  func(f string, a ...any) { fmt.Printf(f, a...) },
	}
}

// sinks returns the current sinks, tolerating a partially installed Output.
func sinks() Output {
	outputMu.RLock()
	o := output
	outputMu.RUnlock()
	if o.Println == nil || o.Printf == nil {
		return defaultOutput()
	}
	return o
}

// New resolves mode ("auto", "true" or "false") to a dialog backend. It never
// returns a nil Dialog.
func New(mode string) Resolution {
	if mode == ModeTrue || mode == ModeAuto {
		if availableOnce() {
			return Resolution{Dialog: zenityDialog{}, Name: zenityDialog{}.Name()}
		}
		// mode=true never aborts the launcher: the library only reports
		// unavailable on Linux/BSD when qarma/zenity/matedialog are missing
		// from PATH, and someone who explicitly asked for graphical dialogs
		// must not be locked out by that.
		if mode == ModeAuto {
			return consoleResolution("")
		}
		return consoleResolution("Graphical dialogs are not available in this system, using the console instead.")
	}
	return consoleResolution("")
}

// consoleResolution builds the console fallback, optionally explaining why the
// configured backend could not be honoured.
func consoleResolution(reason string) Resolution {
	d := consoleDialog{}
	return Resolution{Dialog: d, Name: d.Name(), Reason: reason}
}

// Set installs the dialog for the rest of the session. Set is safe to call
// from tests; RunRoot resets it after teardown.
func Set(d Dialog) {
	active.Store(&d)
}

// Active returns the installed dialog, or a console dialog when Set has not
// been called, so the prompts keep working on early call paths and in tests.
func Active() Dialog {
	if d := active.Load(); d != nil && *d != nil {
		return *d
	}
	return consoleDialog{}
}

// Reset removes the installed dialog.
func Reset() { active.Store(nil) }

// SetOutput installs the sinks used by the console dialog.
func SetOutput(o Output) {
	outputMu.Lock()
	output = o
	outputMu.Unlock()
}
