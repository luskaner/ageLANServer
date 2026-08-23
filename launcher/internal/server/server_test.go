package server

import (
	"net"
	"testing"
)

func TestCalculateBroadcastIPv4(t *testing.T) {
	ip := net.IPv4(192, 168, 10, 23).To4()
	mask := net.IPv4Mask(255, 255, 255, 0)
	got := calculateBroadcastIPv4(ip, mask)
	if !got.Equal(net.IPv4(192, 168, 10, 255).To4()) {
		t.Fatalf("broadcast = %v", got)
	}
}

func TestCalculateBroadcastIPv4HostMask(t *testing.T) {
	ip := net.IPv4(10, 1, 2, 3).To4()
	got := calculateBroadcastIPv4(ip, net.IPv4Mask(255, 255, 255, 255))
	if !got.Equal(ip) {
		t.Fatalf("/32 broadcast = %v, want the ip itself", got)
	}
}
