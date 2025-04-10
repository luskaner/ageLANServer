package relationship

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
	"net/http"
	"strconv"
)

func ChangePresence(users *models.MainUsers, user *models.MainUser, presence int32) {
	user.SetPresence(presence)
	profileInfo := i.A{user.GetProfileInfo(true)}
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
	sess := middleware.Session(r)
	users := models.G(r).Users()
	u, ok := users.GetUserById(sess.GetUserId())
	if ok {
		ChangePresence(users, u, int32(presence))
		i.JSON(&w, i.A{0})
	} else {
		i.JSON(&w, i.A{2})
	}
}
