package account

import (
	"net/http"
	i "server/internal"
	"server/middleware"
	"server/models"
	"strings"
)

func FindProfiles(w http.ResponseWriter, r *http.Request) {
	name := strings.ToLower(r.URL.Query().Get("name"))
	if len(name) < 1 {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	sess, _ := middleware.Session(r)
	u := sess.GetUser()
	profileInfo := models.GetProfileInfo(true, func(currentUser *models.User) bool {
		if u == currentUser {
			return false
		}
		return strings.Contains(strings.ToLower(currentUser.GetAlias()), name)
	})
	i.JSON(&w, i.A{0, profileInfo})
}
