package advertisement

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/middleware"
	"github.com/luskaner/aoe2DELanServer/server/models"
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
	adv, ok := models.GetAdvertisement(int32(advId))
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	currentUser := sess.GetUser()
	_, isPeer := adv.GetPeer(currentUser)
	if !isPeer {
		i.JSON(&w, i.A{2})
		return
	}
	adv.RemovePeer(currentUser)
	i.JSON(&w,
		i.A{0},
	)
}
