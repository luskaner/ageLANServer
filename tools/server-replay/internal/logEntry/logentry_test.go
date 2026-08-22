package logEntry

import (
	"crypto/sha512"
	"net"
	"testing"
	"time"
)

type stubEntry struct {
	uptime   time.Duration
	replayed *[]int
	checked  *[]int
	id       int
}

func (s *stubEntry) Uptime() time.Duration { return s.uptime }

func (s *stubEntry) String() string { return "stub" }

func (s *stubEntry) Replay(net.IP) {
	if s.replayed != nil {
		*s.replayed = append(*s.replayed, s.id)
	}
}

func (s *stubEntry) CheckResponse() {
	if s.checked != nil {
		*s.checked = append(*s.checked, s.id)
	}
}

func resetEntries() { entries = nil }

// Regression: Replay indexed entries[0] and panicked when the decoded log had
// no entries (e.g. an empty server communication log).
func TestReplayWithoutEntriesDoesNotPanic(t *testing.T) {
	resetEntries()
	defer resetEntries()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked: %v", r)
		}
	}()
	Replay(net.ParseIP("127.0.0.1"), true)
}

func TestReplayRunsInUptimeOrderAndChecksAfterwards(t *testing.T) {
	resetEntries()
	defer resetEntries()

	var replayed, checked []int
	Add(&stubEntry{id: 3, uptime: 300, replayed: &replayed, checked: &checked})
	Add(&stubEntry{id: 1, uptime: 100, replayed: &replayed, checked: &checked})
	Add(&stubEntry{id: 2, uptime: 200, replayed: &replayed, checked: &checked})
	// Sorting is the decoder's job (Decode -> Sort -> Replay).
	Sort()

	Replay(net.ParseIP("127.0.0.1"), true)

	if len(replayed) != 3 || replayed[0] != 1 || replayed[1] != 2 || replayed[2] != 3 {
		t.Fatalf("replay order = %v", replayed)
	}
	if len(checked) != 3 || checked[0] != 1 || checked[1] != 2 || checked[2] != 3 {
		t.Fatalf("check order = %v", checked)
	}
}

func TestSameBody(t *testing.T) {
	body := []byte(`{"a":1}`)
	hash := sha512.Sum512(body)
	if !SameBody(hash, body) {
		t.Fatal("matching hash rejected")
	}
	if SameBody([64]byte{1}, body) {
		t.Fatal("non-matching hash accepted")
	}
	if !SameBody([64]byte{}, nil) {
		t.Fatal("empty body must match the zero hash")
	}
}
