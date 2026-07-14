package hosts

import (
	"net"
	"strings"
	"testing"
)

func TestHostMappingsSetGetDelete(t *testing.T) {
	m := make(HostMappings)
	ip := net.ParseIP("127.0.0.1")
	m.Set("example.com", ip)

	got, ok := m.Get("example.com")
	if !ok {
		t.Fatalf("expected Get to find the host")
	}
	if !got.Equal(ip) {
		t.Fatalf("Get() = %v, want %v", got, ip)
	}

	m.Delete("example.com")
	if _, ok := m.Get("example.com"); ok {
		t.Fatalf("expected host to be deleted")
	}
}

func TestHostMappingsGetMissing(t *testing.T) {
	m := make(HostMappings)
	if _, ok := m.Get("missing.com"); ok {
		t.Fatalf("expected ok=false for missing host")
	}
}

func TestHostMappingsString(t *testing.T) {
	m := make(HostMappings)
	m.Set("example.com", net.ParseIP("127.0.0.1"))
	got := m.String("\n")
	want := "\n127.0.0.1\texample.com " + string(commentMarker) + marking + "\n"
	if got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestHostMappingsStringMultiple(t *testing.T) {
	m := make(HostMappings)
	m.Set("a.example.com", net.ParseIP("127.0.0.1"))
	m.Set("b.example.com", net.ParseIP("127.0.0.2"))
	got := m.String("\n")
	if !strings.Contains(got, "a.example.com") || !strings.Contains(got, "b.example.com") {
		t.Fatalf("String() = %q, expected both hosts", got)
	}
	// Every mapping line carries the ownership marking.
	if strings.Count(got, marking) != 2 {
		t.Fatalf("expected marking on each of the 2 lines, got %q", got)
	}
}
