package shared

type AdvertisementBaseRequest struct {
	Id                int32  `schema:"advertisementid"`
	AppBinaryChecksum uint32 `schema:"appbinarychecksum"`
	DataChecksum      uint32 `schema:"datachecksum"`
	ModDllFile        string `schema:"moddllfile"`
	ModDllChecksum    uint32 `schema:"moddllchecksum"`
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
	ModDllFile        string `schema:"moddllfile"`
	ModDllChecksum    uint32 `schema:"moddllchecksum"`
	ModName           string `schema:"modname"`
	ModVersion        string `schema:"modversion"`
	VersionFlags      uint32 `schema:"versionFlags"`
	Description       string `schema:"description"`
	AutomatchPollId   int32  `schema:"automatchPoll_id"`
	MapName           string `schema:"mapname"`
	HostId            int32  `schema:"hostid"`
	Observable        bool   `schema:"isObservable"`
	ObserverDelay     bool   `schema:"observerDelay"`
	ObserverPassword  string `schema:"observerPassword"`
	Password          string `schema:"password"`
	Passworded        bool   `schema:"passworded"`
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
	Observable        bool   `schema:"isObservable"`
	ObserverDelay     bool   `schema:"observerDelay"`
	ObserverPassword  string `schema:"observerPassword"`
	Password          string `schema:"password"`
	Passworded        bool   `schema:"passworded"`
	ModDllFile        string `schema:"moddllfile"`
	ModDllChecksum    uint32 `schema:"moddllchecksum"`
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
