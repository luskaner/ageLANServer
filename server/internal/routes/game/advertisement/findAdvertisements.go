package advertisement

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"

	"net/http"
)

type searchQuery struct {
	AppBinaryChecksum int32  `schema:"appBinaryChecksum"`
	DataChecksum      int32  `schema:"dataChecksum"`
	MatchType         uint8  `schema:"matchType_id"`
	ModDllFile        string `schema:"modDLLFile"`
	ModDllChecksum    int32  `schema:"modDLLChecksum"`
	ModName           string `schema:"modName"`
	ModVersion        string `schema:"modVersion"`
	VersionFlags      uint32 `schema:"versionFlags"`
}

type wanQuery struct {
	Length int `schema:"count"`
	Offset int `schema:"start"`
}

func findAdvertisements(w http.ResponseWriter, r *http.Request, length int, offset int, ongoing bool, lanRegions map[string]struct{}, extraCheck func(*models.MainAdvertisement) bool) {
	var q searchQuery
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	game := models.G(r)
	title := game.Title()
	sess := middleware.Session(r)
	currentUserId := sess.GetUserId()
	var battleServers *models.MainBattleServers
	if len(lanRegions) == 0 {
		battleServers = game.BattleServers()
	}
	advs := game.Advertisements().LockedFindAdvertisementsEncoded(title, length, offset, true, func(adv *models.MainAdvertisement) bool {
		peers := adv.GetPeers()
		_, isPeer := peers.Load(currentUserId)
		var matchesBattleServer bool
		if battleServers == nil {
			_, matchesBattleServer = lanRegions[adv.GetRelayRegion()]
		} else {
			_, matchesBattleServer = battleServers.Get(adv.GetRelayRegion())
		}
		return adv.UnsafeGetJoinable() != ongoing &&
			!isPeer &&
			adv.UnsafeGetVisible() != ongoing &&
			adv.UnsafeGetAppBinaryChecksum() == q.AppBinaryChecksum &&
			adv.UnsafeGetDataChecksum() == q.DataChecksum &&
			adv.UnsafeGetMatchType() == q.MatchType &&
			adv.UnsafeGetModDllFile() == q.ModDllFile &&
			adv.UnsafeGetModDllChecksum() == q.ModDllChecksum &&
			adv.UnsafeGetModName() == q.ModName &&
			adv.UnsafeGetModVersion() == q.ModVersion &&
			adv.UnsafeGetVersionFlags() == q.VersionFlags &&
			matchesBattleServer &&
			(extraCheck == nil || extraCheck(adv))
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

func FindAdvertisements(w http.ResponseWriter, r *http.Request) {
	var q wanQuery
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	findAdvertisements(w, r, q.Length, q.Offset, false, nil, nil)
}
