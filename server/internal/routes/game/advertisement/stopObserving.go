package advertisement

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
)

func StopObserving(w http.ResponseWriter, r *http.Request) {
	var a shared.AdvertisementId
	if err := i.Bind(r, &a); err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	game := models.G(r)
	advertisements := game.Advertisements()
	adv, found := advertisements.GetAdvertisement(a.AdvertisementId)
	if !found {
		i.JSON(&w, i.A{0})
		return
	}
	sess := middleware.SessionOrPanic(r)
	adv.StopObserving(sess.GetUserId())
	i.JSON(&w, i.A{0})
}
