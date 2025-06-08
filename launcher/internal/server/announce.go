package server

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
)

const AnnounceVersionLength = 1
const AnnounceIdLength = 16
const AnnounceVersionSupportedLatest = common.AnnounceVersion2
const AnnounceListen = common.AnnouncePeriod + common.AnnouncePeriod/2

type AnnounceMessage struct {
	// Data is nil for v1.9.0 and higher. Mantained for compatibility for v1.7.3 - v1.8.2
	Data    interface{}
	Version byte
	Ips     mapset.Set[string]
}

type AnnounceMessageDataSupportedLatest = common.AnnounceMessageData002
