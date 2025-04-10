package advertisement

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
	"strconv"
)

func Leave(w http.ResponseWriter, r *http.Request) {
	sess := middleware.Session(r)
	advStr := r.PostFormValue("advertisementid")
	advId64, err := strconv.ParseInt(advStr, 10, 32)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	game := models.G(r)
	advertisements := game.Advertisements()
	advId := int32(advId64)
	var success bool
	advertisements.WithWriteLock(advId, func() {
		adv, ok := advertisements.GetAdvertisement(advId)
		if !ok {
			return
		}
		success = advertisements.UnsafeRemovePeer(adv.GetId(), sess.GetUserId())
	})
	if success {
		i.JSON(&w, i.A{0})
	} else {
		i.JSON(&w, i.A{2})
	}
}
