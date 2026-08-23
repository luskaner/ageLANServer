package cmdUtils

import (
	"strings"
	"testing"
)

func TestSelectServerIndexAutoSelectSingle(t *testing.T) {
	if got := selectServerIndex(1, true, strings.NewReader("")); got != 0 {
		t.Fatalf("auto-select single = %d, want 0", got)
	}
}

func TestSelectServerIndexValidInput(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  int
	}{
		{"1\n", 0},
		{"2\n", 1},
		{"3\n", 2},
	} {
		reader := strings.NewReader(tc.input)
		if got := selectServerIndex(3, false, reader); got != tc.want {
			t.Errorf("input %q: index = %d, want %d", tc.input, got, tc.want)
		}
	}
}

// Regression: out-of-range input used to share the same `continue` as read
// errors; now it retries but a broken reader must give up immediately.
func TestSelectServerIndexInvalidThenValidRetries(t *testing.T) {
	reader := strings.NewReader("99\n0\n2\n")
	if got := selectServerIndex(3, false, reader); got != 1 {
		t.Fatalf("index = %d, want 1 (0-based after retry)", got)
	}
}

// Regression: on stdin EOF the old loop printed the list forever because it
// `continue`d on every Scan error. It must return -1 so the caller can fall
// back to starting its own server.
func TestSelectServerIndexEOFReturnsNegative(t *testing.T) {
	if got := selectServerIndex(3, false, strings.NewReader("")); got != -1 {
		t.Fatalf("EOF = %d, want -1", got)
	}
}
