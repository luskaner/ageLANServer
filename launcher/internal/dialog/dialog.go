// Package dialog abstracts the launcher's interactive prompts so they can be
// answered either in a graphical window or in the console.
//
// It must not import cmdUtils or cmdUtils/logger: cmdUtils needs dialog, and
// cmdUtils/logger needs dialog too, so console output goes through the
// injectable Output sinks instead.
package dialog

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// Mode values accepted by Config.Dialog and --dialog.
const (
	ModeAuto  = "auto"
	ModeTrue  = "true"
	ModeFalse = "false"
)

// ServerCandidate is one discovered server offered to the user.
type ServerCandidate struct {
	// Description is the full line, with every discovered IP and hostname. The
	// console shows it as is.
	Description string
	// Label is a compact summary for backends with little horizontal room.
	// Graphical backends use this so the decisive part (address, latency,
	// version) survives instead of being clipped off the end.
	Label string
}

// label returns the compact summary, falling back to the full description when
// no compact one was provided.
func (c ServerCandidate) label() string {
	if c.Label == "" {
		return c.Description
	}
	return c.Label
}

// Dialog is the set of interactive prompts the launcher needs.
type Dialog interface {
	// Name returns the backend identifier recorded in the log file.
	Name() string

	// SelectServer asks which of the discovered servers to use. servers
	// holds the candidates already sorted by latency. It returns the 0-based
	// index into servers.
	// ok is false when the user declined to pick one, in which case the
	// caller must fall back to its own default (start its own server).
	// stdin is only read by the console implementation.
	SelectServer(servers []ServerCandidate, stdin io.Reader) (index int, ok bool)

	// ListCandidates shows the candidate list without asking anything. The
	// console backend prints it, exactly as SelectServer does before its
	// prompt, so paths that answer on their own still leave the console user
	// with a trace of what was considered. Backends that render the list in
	// their own window leave it as a no-op.
	ListCandidates(servers []ServerCandidate)

	// ConfirmStartServer asks whether to go ahead and start the server.
	// It returns false only when the user actively declined.
	ConfirmStartServer(text string, stdin io.Reader) bool
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
