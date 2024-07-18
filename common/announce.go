package common

import mapset "github.com/deckarep/golang-set/v2"

const AnnouncePort = 31978
const AnnounceHeader = Name
const AnnounceVersionLength = 1
const AnnounceIdLength = 16

const AnnounceVersion0 = 0

const AnnounceVersionLatest = AnnounceVersion0

// AnnounceMessageData000 Empty interface to be used as a placeholder for the message type
type AnnounceMessageData000 struct {
}

type AnnounceMessage struct {
	Data    interface{}
	Version byte
	Ips     mapset.Set[string]
}
