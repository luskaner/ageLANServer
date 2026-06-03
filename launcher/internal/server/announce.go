package server

import (
	"net/netip"

	mapset "github.com/deckarep/golang-set/v2"
)

const AnnounceIdLength = 16

type AnnounceMessage struct {
	IpAddrs mapset.Set[netip.Addr]
}
