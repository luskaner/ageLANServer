package hosts

import (
	"strings"
	"testing"
)

func mustParse(t *testing.T, s string) Line {
	t.Helper()
	ok, _, l := ParseLine(s, true)
	if !ok {
		t.Fatalf("ParseLine(%q) failed", s)
	}
	return l
}

func TestLineString(t *testing.T) {
	l := mustParse(t, "127.0.0.1 example.com")
	if got := l.String(); got != "127.0.0.1\texample.com" {
		t.Fatalf("String() = %q, want %q", got, "127.0.0.1\texample.com")
	}
}

func TestLineOwnFalseByDefault(t *testing.T) {
	l := mustParse(t, "127.0.0.1 example.com")
	if l.Own() {
		t.Fatalf("expected Own()=false for unmarked line")
	}
}

func TestLineWithOwnMarking(t *testing.T) {
	l := mustParse(t, "127.0.0.1 example.com").WithOwnMarking()
	if !l.Own() {
		t.Fatalf("expected Own()=true after WithOwnMarking()")
	}
	if !strings.Contains(l.String(), string(commentMarker)+marking) {
		t.Fatalf("String() = %q, expected to contain marking %q", l.String(), marking)
	}
	// Idempotent: applying twice should not add a second marking.
	l2 := l.WithOwnMarking()
	if strings.Count(l2.String(), marking) != 1 {
		t.Fatalf("marking should appear exactly once, got %q", l2.String())
	}
}

func TestLineWithoutOwnMarking(t *testing.T) {
	l := mustParse(t, "127.0.0.1 example.com").WithOwnMarking().WithoutOwnMarking()
	if l.Own() {
		t.Fatalf("expected Own()=false after WithoutOwnMarking()")
	}
	if strings.Contains(l.String(), marking) {
		t.Fatalf("String() = %q should not contain marking", l.String())
	}
}

func TestLineOnlyComments(t *testing.T) {
	l := mustParse(t, "# only a comment")
	if !l.OnlyComments() {
		t.Fatalf("expected OnlyComments()=true")
	}
	host := mustParse(t, "127.0.0.1 example.com")
	if host.OnlyComments() {
		t.Fatalf("expected OnlyComments()=false for a host line")
	}
}

func TestLineCommentedThenUncommented(t *testing.T) {
	orig := mustParse(t, "127.0.0.1 example.com")
	ok, commented := orig.Commented()
	if !ok {
		t.Fatalf("Commented() returned ok=false")
	}
	if !commented.OnlyComments() {
		t.Fatalf("commented line should be comment-only")
	}
	if got := commented.Uncommented(); got != orig.String() {
		t.Fatalf("Uncommented() = %q, want %q", got, orig.String())
	}
}
