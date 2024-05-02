package relationship

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/middleware"
	"net/http"
	"strconv"
)

func SetPresence(w http.ResponseWriter, r *http.Request) {
	presenceId := r.PostFormValue("presence_id")
	if presenceId == "" {
		i.JSON(&w, i.A{2})
		return
	}
	presence, err := strconv.Atoi(presenceId)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	sess, _ := middleware.Session(r)
	sess.GetUser().SetPresence(int8(presence))
	i.JSON(&w, i.A{0})
}
