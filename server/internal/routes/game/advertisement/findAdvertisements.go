package advertisement

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

type searchQuery struct {
	AppBinaryChecksum int32 `schema:"appBinaryChecksum"`
	DataChecksum      int32 `schema:"dataChecksum"`
	// AoE3 does not send match type when searching for observable games
	MatchType      *uint8 `schema:"matchType_id"`
	ModDllFile     string `schema:"modDLLFile"`
	ModDllChecksum int32  `schema:"modDLLChecksum"`
	ModName        string `schema:"modName"`
	ModVersion     string `schema:"modVersion"`
	VersionFlags   uint32 `schema:"versionFlags"`
}

type wanQuery struct {
	Length int `schema:"count"`
	Offset int `schema:"start"`
}

func findAdvResp(errorCode int, advs i.A) i.A {
	resp := getAdvResp(errorCode, advs)
	resp = append(resp, i.A{})
	return resp
}

func findAdvertisements(w http.ResponseWriter, r *http.Request, length int, offset int, ongoing bool, lanRegions map[string]struct{}, extraCheck func(models.Advertisement) bool) {
	var q searchQuery
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, findAdvResp(2, i.A{}))
		return
	}
	game := models.G(r)
	title := game.Title()
	sess := models.SessionOrPanic(r)
	currentUserId := sess.GetUserId()
	var battleServers models.BattleServers
	if len(lanRegions) == 0 {
		battleServers = game.BattleServers()
	}
	var tagsCheck func(models.Advertisement) bool
	if battleServers != nil && (title == common.GameAoE2 || title == common.GameAoM || title == common.GameAoE4) {
		ok, numericTags, stringTags := parseTags(r)
		if ok {
			tagsCheck = func(adv models.Advertisement) bool {
				return adv.UnsafeMatchesTags(numericTags, stringTags)
			}
		}
	}
	advs := game.Advertisements().LockedFindAdvertisementsEncoded(title, length, offset, true, func(adv models.Advertisement) bool {
		peers := adv.GetPeers()
		_, isPeer := peers.Load(currentUserId)
		var matchesBattleServer bool
		if battleServers == nil {
			_, matchesBattleServer = lanRegions[adv.GetRelayRegion()]
		} else {
			_, matchesBattleServer = battleServers.Get(adv.GetRelayRegion())
		}
		return !isPeer &&
			(tagsCheck == nil || tagsCheck(adv)) &&
			(adv.UnsafeGetJoinable() != ongoing || adv.UnsafeGetVisible() != ongoing) &&
			adv.UnsafeGetAppBinaryChecksum() == q.AppBinaryChecksum &&
			adv.UnsafeGetDataChecksum() == q.DataChecksum &&
			(q.MatchType == nil || adv.UnsafeGetMatchType() == *q.MatchType) &&
			adv.UnsafeGetModDllFile() == q.ModDllFile &&
			adv.UnsafeGetModDllChecksum() == q.ModDllChecksum &&
			adv.UnsafeGetModName() == q.ModName &&
			adv.UnsafeGetModVersion() == q.ModVersion &&
			adv.UnsafeGetVersionFlags() == q.VersionFlags &&
			matchesBattleServer &&
			(extraCheck == nil || extraCheck(adv))
	})
	i.JSON(&w, findAdvResp(0, advs))
}

func FindAdvertisements(w http.ResponseWriter, r *http.Request) {
	var q wanQuery
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, findAdvResp(2, i.A{}))
		return
	}
	findAdvertisements(w, r, q.Length, q.Offset, false, nil, nil)
}
