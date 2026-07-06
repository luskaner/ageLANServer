package advertisement

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common/game"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
)

func updateReturnError(gameId string, w *http.ResponseWriter) {
	response := i.A{2}
	if gameId != game.AoE1 {
		response = append(response, i.A{})
	}
	i.JSON(w, response)
}

func Update(w http.ResponseWriter, r *http.Request) {
	g := models.G(r)
	gameTitle := g.Title()

	var q shared.AdvertisementUpdateRequest
	if err := i.Bind(r, &q); err != nil {
		updateReturnError(gameTitle, &w)
		return
	}
	advertisements := g.Advertisements()
	battleServers := g.BattleServers()
	var response i.A
	var ok bool
	advertisements.WithWriteLock(q.Id, func() {
		var adv models.Advertisement
		adv, ok = advertisements.GetAdvertisement(q.Id)
		if !ok {
			return
		}

		if gameTitle == game.AoE1 || gameTitle == game.AoE3 {
			q.Joinable = true
		}
		adv.UnsafeUpdate(&q)
		if gameTitle == game.AoE1 || gameTitle == game.AoE3 {
			adv.UnsafeUpdatePlatformSessionId(q.PsnSessionId)
		}

		if gameTitle == game.AoE2 || gameTitle == game.AoM || gameTitle == game.AoE4 {
			response = adv.UnsafeEncode(gameTitle, battleServers)
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
