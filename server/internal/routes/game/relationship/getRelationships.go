package relationship

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"net/http"
)

func GetRelationships(w http.ResponseWriter, r *http.Request) {
	// As we don't have knowledge of friends, nor it is supposed to be many players on the server
	// just return all online users as if they were friends
	sess, _ := middleware.Session(r)
	game := models.G(r)
	currentUser, _ := game.Users().GetUserById(sess.GetUserId())
	profileInfo := game.Users().GetProfileInfo(true, func(u *models.MainUser) bool {
		return u != currentUser && u.GetPresence() > 0
	})
	i.JSON(&w, i.A{0, i.A{}, i.A{}, i.A{}, i.A{}, profileInfo, i.A{}, i.A{}})
}
