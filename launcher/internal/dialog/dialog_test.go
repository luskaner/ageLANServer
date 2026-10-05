package dialog

import (
	"sync"
	"testing"
)

func TestNewResolvesMode(t *testing.T) {
	origAvailableOnce := availableOnce
	t.Cleanup(func() { availableOnce = origAvailableOnce })

	for _, tc := range []struct {
		mode       string
		available  bool
		wantName   string
		wantReason string
	}{
		{ModeFalse, false, "console", ""},
		{ModeFalse, true, "console", ""},
		{ModeTrue, true, "zenity", ""},
		{ModeTrue, false, "console", "Graphical dialogs are not available in this system, using the console instead."},
		{ModeAuto, true, "zenity", ""},
		{ModeAuto, false, "console", ""},
	} {
		available := tc.available
		availableOnce = sync.OnceValue(func() bool { return available })
		res := New(tc.mode)
		if res.Name != tc.wantName {
			t.Errorf("mode %q available %t: name = %q, want %q", tc.mode, tc.available, res.Name, tc.wantName)
		}
		if res.Reason != tc.wantReason {
			t.Errorf("mode %q available %t: reason = %q, want %q", tc.mode, tc.available, res.Reason, tc.wantReason)
		}
		if res.Dialog == nil {
			t.Errorf("mode %q available %t: nil dialog", tc.mode, tc.available)
		} else if res.Dialog.Name() != tc.wantName {
			t.Errorf("mode %q available %t: dialog.Name() = %q, want %q", tc.mode, tc.available, res.Dialog.Name(), tc.wantName)
		}
	}
}

// Regression: an unvalidated mode must still resolve to a usable backend, the
// console, instead of a nil Dialog that would panic on the first prompt.
func TestNewUnknownModeFallsBackToConsole(t *testing.T) {
	origAvailableOnce := availableOnce
	t.Cleanup(func() { availableOnce = origAvailableOnce })
	availableOnce = sync.OnceValue(func() bool { return false })

	res := New("xxx")
	if res.Name != "console" || res.Dialog == nil {
		t.Fatalf("unknown mode = %+v, want the console backend", res)
	}
}

func TestActiveDefaultsToConsole(t *testing.T) {
	Reset()
	if got := Active().Name(); got != "console" {
		t.Fatalf("Active() with no Set = %q, want %q", got, "console")
	}
}

func TestSetAndReset(t *testing.T) {
	t.Cleanup(Reset)
	Set(consoleDialog{})
	if got := Active().Name(); got != "console" {
		t.Fatalf("Active() = %q, want %q", got, "console")
	}
	Reset()
	if got := Active().Name(); got != "console" {
		t.Fatalf("Active() after Reset = %q, want %q", got, "console")
	}
}

// Regression: Set(nil) used to hand back a nil Dialog through the stored
// pointer, panicking on the first prompt.
func TestSetNilFallsBackToConsole(t *testing.T) {
	t.Cleanup(Reset)
	Set(nil)
	if got := Active().Name(); got != "console" {
		t.Fatalf("Active() after Set(nil) = %q, want %q", got, "console")
	}
}

func TestIndexFromItem(t *testing.T) {
	for _, tc := range []struct {
		item   string
		want   int
		wantOK bool
	}{
		{item: "1. a", want: 0, wantOK: true},
		{item: "3. x", want: 2, wantOK: true},
		{item: "12. a - 5 ms (1)", want: 11, wantOK: true},
		{item: "x", wantOK: false},
		{item: ". a", wantOK: false},
		{item: "0. a", wantOK: false},
		{item: "a. b", wantOK: false},
		{item: "", wantOK: false},
	} {
		got, ok := indexFromItem(tc.item)
		if ok != tc.wantOK || got != tc.want {
			t.Errorf("indexFromItem(%q) = (%d, %t), want (%d, %t)", tc.item, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestSetOutput(t *testing.T) {
	t.Cleanup(func() { SetOutput(defaultOutput()) })
	var printed, formatted []any
	SetOutput(Output{
		Println: func(a ...any) { printed = a },
		Printf:  func(f string, a ...any) { formatted = append(formatted, f, a) },
	})
	o := sinks()
	o.Println("hello")
	o.Printf("%d\n", 1)
	if len(printed) != 1 || printed[0] != "hello" {
		t.Errorf("Println sink got %v, want [hello]", printed)
	}
	if len(formatted) != 2 || formatted[0] != "%d\n" {
		t.Errorf("Printf sink got %v, want [%%d 1]", formatted)
	}
}

// A partially installed Output must not nil-panic the prompts.
func TestSetOutputPartialKeepsDefaults(t *testing.T) {
	t.Cleanup(func() { SetOutput(defaultOutput()) })
	SetOutput(Output{Println: func(...any) {}})
	o := sinks()
	if o.Printf == nil {
		t.Fatal("Printf sink is nil after a partial SetOutput")
	}
	o.Printf("")
}
