package relationship

import (
	"net/http"
	i "server/internal"
	"server/middleware"
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
