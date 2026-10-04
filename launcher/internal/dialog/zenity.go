package dialog

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/ncruces/zenity"
)

var (
	zenityListFn     = zenity.List
	zenityQuestionFn = zenity.Question
)

// availableOnce probes the system only once per process: the answer cannot
// change while the launcher runs, and IsAvailable shells out on Linux.
var availableOnce = sync.OnceValue(zenity.IsAvailable)

type zenityDialog struct{}

func (zenityDialog) Name() string { return "zenity" }

// ListCandidates is a no-op: the graphical window renders the list itself.
func (zenityDialog) ListCandidates([]ServerCandidate) {}

func (zenityDialog) SelectServer(servers []ServerCandidate, stdin io.Reader) (int, bool) {
	// The Windows backend lays the list out at a fixed 241 px wide and gives it
	// no horizontal scrollbar, so a full description gets its tail clipped off.
	// The compact Label keeps the address, the latency and the version.
	items := make([]string, len(servers))
	for i, s := range servers {
		items[i] = fmt.Sprintf("%d. %s", i+1, label(s))
	}
	selected, err := zenityListFn(
		// The quotes are back here on purpose: a graphical dialog has no colour
		// to mark a program name with, so the quoting still does that job here.
		// <=30 chars: the text control is a fixed 241 px wide on Windows.
		"Select a 'server':",
		items,
		zenity.Title("Select server"),
		zenity.OKLabel("Connect"),
		// Cancelar significa "usar ningún servidor descubierto", que es
		// exactamente lo que el caller ya hace arrancando el suyo. La etiqueta
		// es corta porque el botón mide 75 px fijos en Windows.
		zenity.CancelLabel("Start own"),
		// Solo tienen efecto en Unix; se ignoran en Windows y macOS.
		zenity.Width(600),
		zenity.Height(400),
	)
	if err != nil {
		if errors.Is(err, zenity.ErrCanceled) {
			// "None of these" is a valid answer here: the caller starts its
			// own server.
			return 0, false
		}
		sinks().Println("Could not show the graphical dialog, using the console instead.")
		sinks().Println("Error message: " + err.Error())
		return consoleDialog{}.SelectServer(servers, stdin)
	}
	index, parsed := indexFromItem(selected)
	if !parsed {
		sinks().Println("Could not read the selection from the graphical dialog, using the console instead.")
		return consoleDialog{}.SelectServer(servers, stdin)
	}
	return index, true
}

func (zenityDialog) ConfirmStartServer(text string, stdin io.Reader) bool {
	err := zenityQuestionFn(text,
		zenity.Title("Start server"),
		zenity.OKLabel("Start"),
		zenity.CancelLabel("Cancel"),
		zenity.QuestionIcon,
		zenity.Width(500), // Solo Unix.
	)
	if err != nil {
		if errors.Is(err, zenity.ErrCanceled) {
			return false
		}
		sinks().Println("Could not show the graphical dialog, using the console instead.")
		sinks().Println("Error message: " + err.Error())
		return consoleDialog{}.ConfirmStartServer(text, stdin)
	}
	return true
}

// indexFromItem reverses the "%d. %s" numbering applied to the list items.
func indexFromItem(item string) (int, bool) {
	i := strings.Index(item, ". ")
	if i <= 0 {
		return 0, false
	}
	n, err := strconv.Atoi(item[:i])
	if err != nil || n < 1 {
		return 0, false
	}
	return n - 1, true
}
