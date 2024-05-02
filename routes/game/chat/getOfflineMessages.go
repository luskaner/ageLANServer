package chat

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/middleware"
	"net/http"
	"strconv"
)

func GetOfflineMessages(w http.ResponseWriter, r *http.Request) {
	// TODO: What even are chat channels? plus the server seems to always return the same thing
	sess, _ := middleware.Session(r)
	i.JSON(&w, i.A{0, i.A{}, i.A{i.A{strconv.Itoa(int(sess.GetUser().GetId())), i.A{}}}, i.A{}, i.A{}, i.A{}})
}
