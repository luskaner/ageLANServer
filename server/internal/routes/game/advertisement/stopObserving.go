package advertisement

import (
	"net/http"
	"strconv"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func StopObserving(w http.ResponseWriter, r *http.Request) {
	advId := r.PostFormValue("advertisementid")
	advIdInt64, err := strconv.ParseInt(advId, 10, 32)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	game := models.G(r)
	advertisements := game.Advertisements()
	adv, found := advertisements.GetAdvertisement(int32(advIdInt64))
	if !found {
		i.JSON(&w, i.A{0})
		return
	}
	sess := middleware.Session(r)
	adv.StopObserving(sess.GetUserId())
	i.JSON(&w, i.A{0})
}
