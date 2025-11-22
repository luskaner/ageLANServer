package server

import (
	"net/netip"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
)

const AnnounceIdLength = 16

type AnnounceMessage struct {
	IpAddrs mapset.Set[netip.Addr]
}

type AnnounceMessageDataSupportedLatest = common.AnnounceMessageData002
