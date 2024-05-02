package shared

type ModDllRequest struct {
	File     string `schema:"moddllfile"`
	Checksum uint32 `schema:"moddllchecksum"`
}

type ObserversRequest struct {
	Enabled  bool   `schema:"isObservable"`
	Delay    bool   `schema:"observerDelay"`
	Password string `schema:"observerPassword"`
}

type PasswordBaseRequest struct {
	Value string `schema:"password"`
}

type PasswordRequest struct {
	PasswordBaseRequest
	Enabled bool `schema:"passworded"`
}

type AdvertisementBaseRequest struct {
	Id                int32  `schema:"advertisementid"`
	AppBinaryChecksum uint32 `schema:"appbinarychecksum"`
	DataChecksum      uint32 `schema:"datachecksum"`
	ModDll            ModDllRequest
	ModName           string `schema:"modname"`
	ModVersion        string `schema:"modversion"`
	Party             int32  `schema:"party"`
	Race              int32  `schema:"race"`
	Team              int32  `schema:"team"`
	VersionFlags      uint32 `schema:"versionFlags"`
}

type AdvertisementUpdateRequest struct {
	Id                int32  `schema:"advertisementid"`
	AppBinaryChecksum uint32 `schema:"appbinarychecksum"`
	DataChecksum      uint32 `schema:"datachecksum"`
	ModDll            ModDllRequest
	ModName           string `schema:"modname"`
	ModVersion        string `schema:"modversion"`
	VersionFlags      uint32 `schema:"versionFlags"`
	Description       string `schema:"description"`
	AutomatchPollId   int32  `schema:"automatchPoll_id"`
	MapName           string `schema:"mapname"`
	HostId            int32  `schema:"hostid"`
	Observers         ObserversRequest
	Password          PasswordRequest
	Visible           bool   `schema:"visible"`
	Joinable          bool   `schema:"joinable"`
	MatchType         uint8  `schema:"matchtype"`
	MaxPlayers        uint8  `schema:"maxplayers"`
	Options           string `schema:"options"`
	SlotInfo          string `schema:"slotinfo"`
	PlatformSessionId uint64 `schema:"psnSessionID"`
	State             int8   `schema:"state"`
}

type AdvertisementHostRequest struct {
	AdvertisementBaseRequest
	Description       string `schema:"description"`
	AutomatchPollId   int32  `schema:"automatchPoll_id"`
	RelayRegion       string `schema:"relayRegion"`
	MapName           string `schema:"mapname"`
	HostId            int32  `schema:"hostid"`
	Observers         ObserversRequest
	Password          PasswordRequest
	Visible           bool   `schema:"visible"`
	StatGroup         int32  `schema:"statgroup"`
	Joinable          bool   `schema:"joinable"`
	MatchType         uint8  `schema:"matchtype"`
	MaxPlayers        uint8  `schema:"maxplayers"`
	Options           string `schema:"options"`
	SlotInfo          string `schema:"slotinfo"`
	PlatformSessionId uint64 `schema:"psnSessionID"`
	State             int8   `schema:"state"`
}
