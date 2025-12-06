package relationship

import (
	"net/http"
	"strconv"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

func ChangePresence(clientLibVersion uint16, users models.Users, user models.User, presence int32) {
	user.SetPresence(presence)
	profileInfo := i.A{user.GetProfileInfo(true, clientLibVersion)}
	for u := range users.GetUserIds() {
		sess, ok := models.GetSessionByUserId(u)
		if ok {
			wss.SendOrStoreMessage(
				sess,
				"PresenceMessage",
				profileInfo,
			)
		}
	}
}

func SetPresence(w http.ResponseWriter, r *http.Request) {
	presenceId := r.PostFormValue("presence_id")
	if presenceId == "" {
		i.JSON(&w, i.A{2})
		return
	}
	presence, err := strconv.ParseInt(presenceId, 10, 8)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	sess := models.SessionOrPanic(r)
	game := models.G(r)
	users := game.Users()
	u, ok := users.GetUserById(sess.GetUserId())
	if ok {
		ChangePresence(sess.GetClientLibVersion(), users, u, int32(presence))
		i.JSON(&w, i.A{0})
	} else {
		i.JSON(&w, i.A{2})
	}
}
