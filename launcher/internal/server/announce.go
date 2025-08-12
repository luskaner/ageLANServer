package server

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"net/netip"
	"time"
)

const AnnounceIdLength = 16
const AnnounceVersionSupportedLatest = common.AnnounceVersion2
const AnnounceQuery = 3 * time.Second

type AnnounceMessage struct {
	IpAddrs mapset.Set[netip.Addr]
}

type AnnounceMessageDataSupportedLatest = common.AnnounceMessageData002
