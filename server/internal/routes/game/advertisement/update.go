package advertisement

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
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
	advertisements := game.Advertisements()
	battleServers := game.BattleServers()
	var response i.A
	var ok bool
	advertisements.WithWriteLock(q.Id, func() {
		var adv models.Advertisement
		adv, ok = advertisements.GetAdvertisement(q.Id)
		if !ok {
			return
		}

		if gameTitle != common.GameAoE2 && gameTitle != common.GameAoM {
			q.Joinable = true
		}
		adv.UnsafeUpdate(&q)
		if gameTitle != common.GameAoE2 && gameTitle != common.GameAoM {
			adv.UnsafeUpdatePlatformSessionId(q.PsnSessionId)
		}

		if gameTitle == common.GameAoE2 || gameTitle == common.GameAoM {
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
