package advertisement

import (
	"encoding/json"
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

func GetAdvertisements(w http.ResponseWriter, r *http.Request) {
	matchIdsStr := r.URL.Query().Get("match_ids")
	var advsIds []int32
	err := json.Unmarshal([]byte(matchIdsStr), &advsIds)
	if err != nil {
		i.JSON(&w, getAdvResp(2, i.A{}))
		return
	}
	game := models.G(r)
	title := game.Title()
	advertisements := game.Advertisements()
	advs := advertisements.LockedFindAdvertisementsEncoded(title, 0, 0, false, func(adv *models.MainAdvertisement) bool {
		return slices.Contains(advsIds, adv.GetId())
	})
	if advs == nil {
		i.JSON(&w, getAdvResp(0, i.A{}))
	} else {
		i.JSON(&w, getAdvResp(0, advs))
	}
}
