package advertisement

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
	"strings"
)

type query struct {
	AppBinaryChecksum int32  `schema:"appBinaryChecksum"`
	DataChecksum      int32  `schema:"dataChecksum"`
	MatchType         uint8  `schema:"matchType_id"`
	ModDllFile        string `schema:"modDLLFile"`
	ModDllChecksum    int32  `schema:"modDLLChecksum"`
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
	lanServerGuids := strings.Split(strings.ReplaceAll(strings.Trim(q.RelayRegions, `[]`), `"`, ``), ",")
	sess := middleware.Session(r)
	lanServerGuidsMap := make(map[string]struct{}, len(lanServerGuids))
	for _, guid := range lanServerGuids {
		lanServerGuidsMap[guid] = struct{}{}
	}
	game := models.G(r)
	title := game.Title()
	currentUserId := sess.GetUserId()
	advs := game.Advertisements().LockedFindAdvertisementsEncoded(title, true, func(adv *models.MainAdvertisement) bool {
		_, relayRegionMatches := lanServerGuidsMap[adv.GetRelayRegion()]
		peers := adv.GetPeers()
		_, isPeer := peers.Load(currentUserId)
		return adv.UnsafeGetJoinable() &&
			adv.UnsafeGetVisible() &&
			!isPeer &&
			adv.UnsafeGetAppBinaryChecksum() == q.AppBinaryChecksum &&
			adv.UnsafeGetDataChecksum() == q.DataChecksum &&
			adv.UnsafeGetMatchType() == q.MatchType &&
			adv.UnsafeGetModDllFile() == q.ModDllFile &&
			adv.UnsafeGetModDllChecksum() == q.ModDllChecksum &&
			adv.UnsafeGetModName() == q.ModName &&
			adv.UnsafeGetModVersion() == q.ModVersion &&
			adv.UnsafeGetVersionFlags() == q.VersionFlags &&
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
