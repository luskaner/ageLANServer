package relationship

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

func ChangePresence(clientLibVersion uint16, sessions models.Sessions, users models.Users, user models.User, presence int32) {
	user.SetPresence(presence)
	profileInfo := i.A{user.GetProfileInfo(true, clientLibVersion)}
	for u := range users.GetUserIds() {
		sess, ok := sessions.GetByUserId(u)
		if ok {
			wss.SendOrStoreMessage(
				sess,
				"PresenceMessage",
				profileInfo,
			)
		}
	}
}

type setPresenceRequest struct {
	PresenceId int32 `schema:"presence_id"`
}

func SetPresence(w http.ResponseWriter, r *http.Request) {
	var req setPresenceRequest
	if err := i.Bind(r, &req); err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	sess := models.SessionOrPanic(r)
	game := models.G(r)
	users := game.Users()
	u, ok := users.GetUserById(sess.GetUserId())
	if ok {
		ChangePresence(sess.GetClientLibVersion(), game.Sessions(), users, u, req.PresenceId)
		i.JSON(&w, i.A{0})
	} else {
		i.JSON(&w, i.A{2})
	}
}
