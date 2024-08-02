package relationship

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/middleware"
	"github.com/luskaner/aoe2DELanServer/server/models"
	"net/http"
)

func GetRelationships(w http.ResponseWriter, r *http.Request) {
	// As we don't have knowledge of friends, nor it is supposed to be many players on the server
	// just return all online users as if they were friends
	sess, _ := middleware.Session(r)
	currentUser := sess.GetUser()
	profileInfo := models.GetProfileInfo(true, func(u *models.User) bool {
		return u != currentUser && u.GetPresence() > 0
	})
	i.JSON(&w, i.A{0, i.A{}, i.A{}, i.A{}, i.A{}, profileInfo, i.A{}, i.A{}})
}
