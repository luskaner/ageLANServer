package chat

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"net/http"
	"strconv"
)

func GetOfflineMessages(w http.ResponseWriter, r *http.Request) {
	// Only AoE3 has offline messages but are not implemented as we would need to store the chat messages and who
	// joined the chat channels
	sess, _ := middleware.Session(r)
	i.JSON(&w, i.A{0, i.A{}, i.A{i.A{strconv.Itoa(int(sess.GetUserId())), i.A{}}}, i.A{}, i.A{}, i.A{}})
}
