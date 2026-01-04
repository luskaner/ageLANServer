package advertisement

import (
	"net/http"
	"slices"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func getAdvResp(errorCode int, advs i.A) i.A {
	return i.A{
		errorCode,
		advs,
	}
}

type getAdvertisementsRequest struct {
	MatchIDs i.Json[[]int32] `schema:"match_ids"`
}

func GetAdvertisements(w http.ResponseWriter, r *http.Request) {
	var req getAdvertisementsRequest
	err := i.Bind(r, &req)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	game := models.G(r)
	title := game.Title()
	advertisements := game.Advertisements()
	advs := advertisements.LockedFindAdvertisementsEncoded(title, 0, 0, false, func(adv models.Advertisement) bool {
		return slices.Contains(req.MatchIDs.Data, adv.GetId())
	})
	if advs == nil {
		i.JSON(&w, getAdvResp(0, i.A{}))
	} else {
		i.JSON(&w, getAdvResp(0, advs))
	}
}
