package advertisement

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

type findAdvQuery struct {
	AppBinaryChecksum int32  `schema:"appBinaryChecksum"`
	DataChecksum      int32  `schema:"dataChecksum"`
	MatchType         uint8  `schema:"matchType_id"`
	ModDllFile        string `schema:"modDLLFile"`
	ModDllChecksum    int32  `schema:"modDLLChecksum"`
	ModName           string `schema:"modName"`
	ModVersion        string `schema:"modVersion"`
	VersionFlags      uint32 `schema:"versionFlags"`
}

func findAdvertisements(w http.ResponseWriter, r *http.Request, game models.Game, extraCheck func(*models.MainAdvertisement) bool) {
	if extraCheck == nil {
		extraCheck = func(_ *models.MainAdvertisement) bool { return true }
	}
	var data findAdvQuery
	if err := i.Bind(r, &data); err != nil {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	sess, _ := middleware.Session(r)
	currentUser, _ := game.Users().GetUserById(sess.GetUserId())
	advs := game.Advertisements().FindAdvertisementsEncoded(game.Title(), func(adv *models.MainAdvertisement) bool {
		_, isPeer := adv.GetPeer(currentUser)
		return adv.GetJoinable() &&
			adv.GetVisible() &&
			!isPeer &&
			adv.GetAppBinaryChecksum() == data.AppBinaryChecksum &&
			adv.GetDataChecksum() == data.DataChecksum &&
			adv.GetMatchType() == data.MatchType &&
			adv.GetModDllFile() == data.ModDllFile &&
			adv.GetModDllChecksum() == data.ModDllChecksum &&
			adv.GetModName() == data.ModName &&
			adv.GetModVersion() == data.ModVersion &&
			adv.GetVersionFlags() == data.VersionFlags &&
			extraCheck(adv)

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
	findAdvertisements(w, r, models.G(r), nil)
}
