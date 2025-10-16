package account

import (
	"net/http"
	"strings"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func FindProfiles(w http.ResponseWriter, r *http.Request) {
	name := strings.ToLower(r.URL.Query().Get("name"))
	if len(name) < 1 {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	game := models.G(r)
	gameTitle := game.Title()
	sess := models.SessionOrPanic(r)
	profileInfo := game.Users().GetProfileInfo(true, func(currentUser *models.MainUser) bool {
		return strings.Contains(strings.ToLower(currentUser.GetAlias()), name)
	}, gameTitle, sess.GetClientLibVersion())
	i.JSON(&w, i.A{0, profileInfo})
}
