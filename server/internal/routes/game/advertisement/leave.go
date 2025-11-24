package advertisement

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"

	"net/http"
)

func Leave(w http.ResponseWriter, r *http.Request) {
	var q shared.AdvertisementId
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	game := models.G(r)
	advertisements := game.Advertisements()
	var success bool
	advertisements.WithWriteLock(q.AdvertisementId, func() {
		adv, ok := advertisements.GetAdvertisement(q.AdvertisementId)
		if !ok {
			return
		}
		sess := models.SessionOrPanic(r)
		success = advertisements.UnsafeRemovePeer(adv.GetId(), sess.GetUserId())
	})
	if success {
		i.JSON(&w, i.A{0})
	} else {
		i.JSON(&w, i.A{2})
	}
}
