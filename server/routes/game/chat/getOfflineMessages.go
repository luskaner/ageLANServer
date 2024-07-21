package chat

import (
	"net/http"
	i "server/internal"
	"server/middleware"
	"strconv"
)

func GetOfflineMessages(w http.ResponseWriter, r *http.Request) {
	// What even are chat channels? plus the server seems to always return the same thing
	sess, _ := middleware.Session(r)
	i.JSON(&w, i.A{0, i.A{}, i.A{i.A{strconv.Itoa(int(sess.GetUser().GetId())), i.A{}}}, i.A{}, i.A{}, i.A{}})
}
