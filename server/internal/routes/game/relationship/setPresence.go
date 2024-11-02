package relationship

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/wss"
	"net/http"
	"strconv"
)

func ChangePresence(users *models.MainUsers, user *models.MainUser, presence int8) {
	user.SetPresence(presence)
	profileInfo := i.A{user.GetProfileInfo(true)}
	for _, u := range users.GetUserIds() {
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
	sess, _ := middleware.Session(r)
	users := models.G(r).Users()
	u, _ := users.GetUserById(sess.GetUserId())
	ChangePresence(users, u, int8(presence))
	i.JSON(&w, i.A{0})
}
