package advertisement

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
	"strconv"
)

func Leave(w http.ResponseWriter, r *http.Request) {
	sess, _ := middleware.Session(r)
	advStr := r.PostFormValue("advertisementid")
	advId, err := strconv.ParseInt(advStr, 10, 32)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	game := models.G(r)
	advertisements := game.Advertisements()
	adv, ok := advertisements.GetAdvertisement(int32(advId))
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	currentUser, _ := game.Users().GetUserById(sess.GetUserId())
	_, isPeer := adv.GetPeer(currentUser)
	if !isPeer {
		i.JSON(&w, i.A{2})
		return
	}
	advertisements.RemovePeer(adv, currentUser)
	i.JSON(&w,
		i.A{0},
	)
}
