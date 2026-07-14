package common

import (
	"net"
	"net/netip"
	"reflect"
	"sort"
	"testing"
)

func TestNetIPToNetIPAddr_IPv4(t *testing.T) {
	addr := NetIPToNetIPAddr(net.ParseIP("192.168.1.10"))
	if !addr.IsValid() {
		t.Fatalf("expected valid addr, got invalid")
	}
	if addr.String() != "192.168.1.10" {
		t.Fatalf("addr = %q, want %q", addr.String(), "192.168.1.10")
	}
	if !addr.Is4() {
		t.Fatalf("expected Is4() to be true")
	}
}

func TestNetIPToNetIPAddr_NonIPv4ReturnsInvalid(t *testing.T) {
	addr := NetIPToNetIPAddr(net.ParseIP("2001:db8::1"))
	if addr.IsValid() {
		t.Fatalf("expected invalid addr for IPv6-only input, got %q", addr.String())
	}
}

func TestNetIPAddrToNetIP_RoundTrip(t *testing.T) {
	orig := net.ParseIP("10.0.0.1")
	addr := netip.MustParseAddr("10.0.0.1")
	got := NetIPAddrToNetIP(addr)
	if !got.Equal(orig) {
		t.Fatalf("got %v, want %v", got, orig)
	}
}

func TestStringSliceToNetIPSlice_FiltersInvalid(t *testing.T) {
	got := StringSliceToNetIPSlice([]string{"1.2.3.4", "not-an-ip", "5.6.7.8", ""})
	if len(got) != 2 {
		t.Fatalf("expected 2 valid IPs, got %d (%v)", len(got), got)
	}
	strs := []string{got[0].String(), got[1].String()}
	sort.Strings(strs)
	want := []string{"1.2.3.4", "5.6.7.8"}
	if !reflect.DeepEqual(strs, want) {
		t.Fatalf("got %v, want %v", strs, want)
	}
}

func TestStringSliceToNetIPSlice_Empty(t *testing.T) {
	got := StringSliceToNetIPSlice(nil)
	if got == nil {
		t.Fatalf("expected non-nil empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestNetIPSliceToNetIPSet(t *testing.T) {
	ips := []net.IP{net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2"), net.ParseIP("1.1.1.1")}
	set := NetIPSliceToNetIPSet(ips)
	if set.Cardinality() != 2 {
		t.Fatalf("expected 2 unique addrs, got %d", set.Cardinality())
	}
	if !set.ContainsOne(netip.MustParseAddr("1.1.1.1")) {
		t.Fatalf("expected set to contain 1.1.1.1")
	}
	if !set.ContainsOne(netip.MustParseAddr("2.2.2.2")) {
		t.Fatalf("expected set to contain 2.2.2.2")
	}
}
