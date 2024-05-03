package advertisement

import (
	"encoding/json"
	"net/http"
	i "server/internal"
	"server/middleware"
	"server/models"
)

type query struct {
	AppBinaryChecksum uint32 `schema:"appBinaryChecksum"`
	DataChecksum      uint32 `schema:"dataChecksum"`
	MatchType         uint8  `schema:"matchType_id"`
	ModDllFile        string `schema:"modDLLFile"`
	ModDllChecksum    uint32 `schema:"modDLLChecksum"`
	ModName           string `schema:"modName"`
	ModVersion        string `schema:"modVersion"`
	VersionFlags      uint32 `schema:"versionFlags"`
	RelayRegions      string `schema:"lanServerGuids"`
}

func GetLanAdvertisements(w http.ResponseWriter, r *http.Request) {
	var q query
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	var lanServerGuids []string
	err := json.Unmarshal([]byte(q.RelayRegions), &lanServerGuids)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	sess, _ := middleware.Session(r)
	lanServerGuidsMap := make(map[string]struct{}, len(lanServerGuids))
	for _, guid := range lanServerGuids {
		lanServerGuidsMap[guid] = struct{}{}
	}
	currentUser := sess.GetUser()
	advs := models.FindAdvertisementsEncoded(func(adv *models.Advertisement) bool {
		_, relayRegionMatches := lanServerGuidsMap[adv.GetRelayRegion()]
		_, isPeer := adv.GetPeer(currentUser)
		return adv.GetJoinable() &&
			adv.GetVisible() &&
			!isPeer &&
			adv.GetAppBinaryChecksum() == q.AppBinaryChecksum &&
			adv.GetDataChecksum() == q.DataChecksum &&
			adv.GetMatchType() == q.MatchType &&
			adv.GetModDllFile() == q.ModDllFile &&
			adv.GetModDllChecksum() == q.ModDllChecksum &&
			adv.GetModName() == q.ModName &&
			adv.GetModVersion() == q.ModVersion &&
			adv.GetVersionFlags() == q.VersionFlags &&
			relayRegionMatches
	})
	if advs == nil {
		i.JSON(&w,
			i.A{0, i.A{}, i.A{}},
		)
	} else {
		i.JSON(&w,
			i.A{0, advs, i.A{}},
		)
	}
}
