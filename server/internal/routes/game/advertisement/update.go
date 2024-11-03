package advertisement

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
	"net/http"
)

func Update(w http.ResponseWriter, r *http.Request) {
	var q shared.AdvertisementUpdateRequest
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	game := models.G(r)
	gameTitle := game.Title()

	advertisements := models.G(r).Advertisements()
	adv, ok := advertisements.GetAdvertisement(q.Id)
	if !ok {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	if gameTitle == common.GameAoE3 {
		q.PlatformSessionId = adv.GetPlatformSessionId()
		q.Joinable = true
	}
	advertisements.Update(adv, &q)

	if gameTitle == common.GameAoE3 {
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
