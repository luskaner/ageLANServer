package common

const AnnouncePort = 31978
const AnnounceMulticastGroup = "239.31.97.8"
const AnnounceHeader = Name
const IdHeader = "X-Id"
const VersionHeader = "X-Version"

//goland:noinspection GoUnusedConst
const (
	// AnnounceVersion0 : 1.1.X - v1.4.X
	AnnounceVersion0 = iota
	// AnnounceVersion1 : 1.5.X - v1.8.X
	AnnounceVersion1
	// AnnounceVersion2 (v1.9.X and higher) no longer sends any extra data, '/test' is supposed to be queried to get it
	AnnounceVersion2
)

// Version 1.0.X no additional data was sent

// AnnounceMessageData000 Empty interface equivalent when no struct was used to allow for a future expansion
// 1.1.X - v1.4.X
type AnnounceMessageData000 struct {
}

// AnnounceMessageData001 adds the games supported as not only age2 is supported
// v1.5.X - v1.8.X
type AnnounceMessageData001 struct {
	GameTitles []string
}

// AnnounceMessageData002 adds the version of the server so the launcher can better determine compatibility
// v1.9.X and higher
type AnnounceMessageData002 struct {
	AnnounceMessageData001
	Version string
}
