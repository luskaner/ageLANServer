package advertisement

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
	"net/http"
)

func updateReturnError(gameId string, w *http.ResponseWriter) {
	response := i.A{2}
	if gameId != common.GameAoE1 {
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
	adv, ok := advertisements.GetAdvertisement(q.Id)
	if !ok {
		updateReturnError(gameTitle, &w)
		return
	}
	if gameTitle != common.GameAoE2 {
		q.PlatformSessionId = adv.GetPlatformSessionId()
		q.Joinable = true
	}
	advertisements.Update(adv, &q)

	if gameTitle != common.GameAoE2 {
		i.JSON(&w,
			i.A{
				0,
			},
		)
	} else {
		i.JSON(&w,
			i.A{
				0,
				adv.Encode(gameTitle),
			},
		)
	}

}
