package advertisement

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
	"net/http"
)

func updateReturnError(gameTitle common.GameTitle, w *http.ResponseWriter) {
	response := i.A{2}
	if gameTitle != common.AoE1 {
		response = append(response, i.A{})
	}
	i.JSON(w, response)
}

func Update(w http.ResponseWriter, r *http.Request) {
	game := models.G(r)
	gameTitle := game.Title()

	var q shared.AdvertisementUpdateRequest
	if err := i.Bind(r, &q); err != nil {
		updateReturnError(gameTitle, &w)
		return
	}

	advertisements := models.G(r).Advertisements()
	var response i.A
	var ok bool
	advertisements.WithWriteLock(q.Id, func() {
		var adv *models.MainAdvertisement
		adv, ok = advertisements.GetAdvertisement(q.Id)
		if !ok {
			return
		}

		if gameTitle != common.AoE2 {
			q.PlatformSessionId = adv.UnsafeGetPlatformSessionId()
			q.Joinable = true
		}
		advertisements.UpdateUnsafe(adv, &q)

		if gameTitle == common.AoE2 {
			response = adv.UnsafeEncode(gameTitle)
		}
		ok = true
	})
	if !ok {
		updateReturnError(gameTitle, &w)
		return
	}
	if response == nil {
		i.JSON(&w,
			i.A{
				0,
			},
		)
	} else {
		i.JSON(&w,
			i.A{
				0,
				response,
			},
		)
	}
}
