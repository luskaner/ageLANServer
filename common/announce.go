package common

import mapset "github.com/deckarep/golang-set/v2"

const AnnouncePort = 31978
const AnnounceHeader = Name
const AnnounceVersionLength = 1
const AnnounceIdLength = 16

const (
	AnnounceVersion0 = iota
	AnnounceVersion1
)

const AnnounceVersionLatest = AnnounceVersion1

// AnnounceMessageData000 Empty interface to be used as a placeholder for the message type
type AnnounceMessageData000 struct {
}

// AnnounceMessageData001 Data structure for the announce version 1
type AnnounceMessageData001 struct {
	GameIds []string
}

type AnnounceMessage struct {
	Data    interface{}
	Version byte
	Ips     mapset.Set[string]
}
