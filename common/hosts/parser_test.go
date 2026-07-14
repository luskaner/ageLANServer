package hosts

import (
	"net"
	"testing"
)

func TestParseLine_HostMapping(t *testing.T) {
	ok, overLimit, l := ParseLine("127.0.0.1 example.com", true)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if overLimit {
		t.Fatalf("expected overLimit=false")
	}
	if l.OnlyComments() {
		t.Fatalf("expected a host mapping, got comment-only line")
	}
	if !l.IP().Equal(net.ParseIP("127.0.0.1")) {
		t.Fatalf("IP() = %v, want 127.0.0.1", l.IP())
	}
	hosts := l.Hosts()
	if len(hosts) != 1 || hosts[0].String() != "example.com" {
		t.Fatalf("Hosts() = %v, want [example.com]", hosts)
	}
}

func TestParseLine_MultipleHosts(t *testing.T) {
	ok, _, l := ParseLine("10.0.0.5 a.example.com b.example.com", true)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if len(l.Hosts()) != 2 {
		t.Fatalf("expected 2 hosts, got %v", l.Hosts())
	}
}

func TestParseLine_CommentOnly(t *testing.T) {
	ok, _, l := ParseLine("# just a comment", true)
	if !ok {
		t.Fatalf("expected ok=true for comment-only line")
	}
	if !l.OnlyComments() {
		t.Fatalf("expected OnlyComments()=true")
	}
}

func TestParseLine_Empty(t *testing.T) {
	ok, _, l := ParseLine("", true)
	if !ok {
		t.Fatalf("expected ok=true for empty line")
	}
	if !l.OnlyComments() {
		t.Fatalf("expected OnlyComments()=true for empty line")
	}
}

func TestParseLine_TrailingComment(t *testing.T) {
	ok, _, l := ParseLine("127.0.0.1 example.com # note", true)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if l.OnlyComments() {
		t.Fatalf("expected a host mapping")
	}
	if len(l.Hosts()) != 1 {
		t.Fatalf("expected 1 host, got %v", l.Hosts())
	}
}

func TestParseLine_MissingHost(t *testing.T) {
	ok, _, _ := ParseLine("127.0.0.1", true)
	if ok {
		t.Fatalf("expected ok=false when there is no host")
	}
}

func TestParseLine_InvalidIP(t *testing.T) {
	ok, _, _ := ParseLine("not-an-ip example.com", true)
	if ok {
		t.Fatalf("expected ok=false when the address is not an IP")
	}
}

func TestParseLine_OwnMarkingRoundTrip(t *testing.T) {
	ok, _, l := ParseLine("127.0.0.1\texample.com #"+marking, true)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if !l.Own() {
		t.Fatalf("expected Own()=true for line marked with %q", marking)
	}
}

func TestParseHost(t *testing.T) {
	ok, host := parseHost("example.com")
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if host.String() != "example.com" {
		t.Fatalf("host = %q, want example.com", host.String())
	}
}

func TestParseHost_RejectsIP(t *testing.T) {
	if ok, _ := parseHost("127.0.0.1"); ok {
		t.Fatalf("expected parseHost to reject a literal IP")
	}
}
