package account

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"net/http"
	"strings"
)

func FindProfiles(w http.ResponseWriter, r *http.Request) {
	name := strings.ToLower(r.URL.Query().Get("name"))
	if len(name) < 1 {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	sess, _ := middleware.Session(r)
	game := models.G(r)
	users := game.Users()
	u, _ := users.GetUserById(sess.GetUserId())
	profileInfo := users.GetProfileInfo(true, func(currentUser *models.MainUser) bool {
		if u == currentUser {
			return false
		}
		return strings.Contains(strings.ToLower(currentUser.GetAlias()), name)
	})
	i.JSON(&w, i.A{0, profileInfo})
}
