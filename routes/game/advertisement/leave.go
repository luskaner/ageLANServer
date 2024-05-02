package advertisement

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/middleware"
	"aoe2DELanServer/models"
	"net/http"
	"strconv"
)

func Leave(w http.ResponseWriter, r *http.Request) {
	sess, _ := middleware.Session(r)
	advStr := r.PostFormValue("advertisementid")
	advId, err := strconv.ParseUint(advStr, 10, 32)
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
