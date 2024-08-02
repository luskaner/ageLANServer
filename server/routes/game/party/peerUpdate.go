package party

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/middleware"
	"github.com/luskaner/aoe2DELanServer/server/models"
	"github.com/luskaner/aoe2DELanServer/server/routes/game/party/shared"
	"net/http"
)

func PeerUpdate(w http.ResponseWriter, r *http.Request) {
	// What about isNonParticipants[]? observers? ai players?
	adv, length, profileIds, raceIds, statGroupIds, teamIds := shared.ParseParameters(r)
	if adv == nil {
		i.JSON(&w, i.A{2})
		return
	}
	sess, _ := middleware.Session(r)
	currentUser := sess.GetUser()
	// Only the host can update peers
	if adv.GetHost().GetUser() != currentUser {
		i.JSON(&w, i.A{2})
		return
	}
	users := make([]*models.User, length)
	for j := 0; j < length; j++ {
		u, ok := models.GetUserById(profileIds[j])
		if !ok || u.GetStatId() != statGroupIds[j] {
			i.JSON(&w, i.A{2})
			return
		}
		users[j] = u
	}
	for j, u := range users {
		adv.UpdatePeer(u, raceIds[j], teamIds[j])
	}
	i.JSON(&w, i.A{0})
}
