package models

import (
	"net"
	"net/http/httptest"
	"testing"

	"github.com/luskaner/ageLANServer/server/internal"
)

func resetNetworkState(t *testing.T) {
	t.Helper()
	oldCanUseInternet := internal.CanUseInternet
	oldPublicIp := publicIp
	oldLocalSubnets := localSubnets
	t.Cleanup(func() {
		internal.CanUseInternet = oldCanUseInternet
		publicIp = oldPublicIp
		localSubnets = oldLocalSubnets
	})
	publicIp = ""
	localSubnets = nil
}

func autoBattleServer() *MainBattleServer {
	bs := &MainBattleServer{}
	bs.SetIPv4("auto")
	return bs
}

func TestCacheNetworkInterfacesNoInternet(t *testing.T) {
	resetNetworkState(t)
	internal.CanUseInternet = false

	CacheNetworkInterfaces("")
	if publicIp != "" {
		t.Fatalf("publicIp = %q, want empty without internet", publicIp)
	}
	if len(localSubnets) != 0 {
		t.Fatalf("localSubnets = %d entries, want none without a public IP", len(localSubnets))
	}
}

func TestCacheNetworkInterfacesAutoOffline(t *testing.T) {
	resetNetworkState(t)
	internal.CanUseInternet = false

	CacheNetworkInterfaces("auto")
	if publicIp != "" {
		t.Fatalf("publicIp = %q, want empty with 'auto' and no internet", publicIp)
	}
	if len(localSubnets) != 0 {
		t.Fatalf("localSubnets = %d entries, want none with 'auto' and no internet", len(localSubnets))
	}
}

func TestCacheNetworkInterfacesExternalIP(t *testing.T) {
	resetNetworkState(t)
	internal.CanUseInternet = false

	CacheNetworkInterfaces("1.2.3.4")
	if publicIp != "1.2.3.4" {
		t.Fatalf("publicIp = %q, want 1.2.3.4 from external IP", publicIp)
	}
	if len(localSubnets) == 0 {
		t.Fatal("localSubnets should be populated when a public IP is available")
	}
}

func TestResolveIPv4OfflineNoPublicFallback(t *testing.T) {
	resetNetworkState(t)
	internal.CanUseInternet = false
	publicIp = "8.8.8.8"
	_, subnet, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	localSubnets = []*net.IPNet{subnet}

	bs := autoBattleServer()
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "9.9.9.9:5555"
	if ip := bs.ResolveIPv4(req); ip != "" {
		t.Fatalf("ResolveIPv4 = %q, want empty (no public IP leak offline)", ip)
	}
}

func TestResolveIPv4OnlinePublicFallback(t *testing.T) {
	resetNetworkState(t)
	internal.CanUseInternet = true
	publicIp = "8.8.8.8"
	_, subnet, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	localSubnets = []*net.IPNet{subnet}

	bs := autoBattleServer()

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "9.9.9.9:5555"
	req.Host = "cdn.example.com"
	if ip := bs.ResolveIPv4(req); ip != "8.8.8.8" {
		t.Fatalf("ResolveIPv4 = %q, want 8.8.8.8 public fallback", ip)
	}

	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "9.9.9.9:5555"
	req.Host = "10.1.2.3"
	if ip := bs.ResolveIPv4(req); ip != "10.1.2.3" {
		t.Fatalf("ResolveIPv4 = %q, want 10.1.2.3 from host", ip)
	}

	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.50:5555"
	if ip := bs.ResolveIPv4(req); ip != "" {
		t.Fatalf("ResolveIPv4 = %q, want empty for LAN peer", ip)
	}
}
