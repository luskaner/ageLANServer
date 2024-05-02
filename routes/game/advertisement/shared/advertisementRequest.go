package shared

type ModDllRequest struct {
	File     string `form:"moddllfile"`
	Checksum uint32 `form:"moddllchecksum"`
}

type ObserversRequest struct {
	Enabled  bool   `form:"isObservable"`
	Delay    bool   `form:"observerDelay"`
	Password string `form:"observerPassword"`
}

type PasswordBaseRequest struct {
	Value string `form:"password"`
}

type PasswordRequest struct {
	PasswordBaseRequest
	Enabled bool `form:"passworded"`
}

type AdvertisementBaseRequest struct {
	Id                int64  `form:"advertisementid"`
	AppBinaryChecksum uint32 `form:"appbinarychecksum"`
	DataChecksum      uint32 `form:"datachecksum"`
	ModDll            ModDllRequest
	ModName           string `form:"modname"`
	ModVersion        string `form:"modversion"`
	Party             int32  `form:"party"`
	Race              int32  `form:"race"`
	Team              int32  `form:"team"`
	VersionFlags      uint32 `form:"versionFlags"`
}

type AdvertisementUpdateRequest struct {
	Id                int64  `form:"advertisementid"`
	AppBinaryChecksum uint32 `form:"appbinarychecksum"`
	DataChecksum      uint32 `form:"datachecksum"`
	ModDll            ModDllRequest
	ModName           string `form:"modname"`
	ModVersion        string `form:"modversion"`
	VersionFlags      uint32 `form:"versionFlags"`
	Description       string `form:"description"`
	AutomatchPollId   int32  `form:"automatchPoll_id"`
	MapName           string `form:"mapname"`
	HostId            int32  `form:"hostid"`
	Observers         ObserversRequest
	Password          PasswordRequest
	Visible           bool   `form:"visible"`
	Joinable          bool   `form:"joinable"`
	MatchType         uint8  `form:"matchtype"`
	MaxPlayers        uint8  `form:"maxplayers"`
	Options           string `form:"options"`
	SlotInfo          string `form:"slotinfo"`
	PlatformSessionId uint64 `form:"psnSessionID"`
	State             int8   `form:"state"`
}

type AdvertisementHostRequest struct {
	AdvertisementBaseRequest
	Description       string `form:"description"`
	AutomatchPollId   int32  `form:"automatchPoll_id"`
	RelayRegion       string `form:"relayRegion"`
	MapName           string `form:"mapname"`
	HostId            int32  `form:"hostid"`
	Observers         ObserversRequest
	Password          PasswordRequest
	Visible           bool   `form:"visible"`
	StatGroup         int32  `form:"statgroup"`
	Joinable          bool   `form:"joinable"`
	MatchType         uint8  `form:"matchtype"`
	MaxPlayers        uint8  `form:"maxplayers"`
	Options           string `form:"options"`
	SlotInfo          string `form:"slotinfo"`
	PlatformSessionId uint64 `form:"psnSessionID"`
	State             int8   `form:"state"`
}
