package party

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/middleware"
	"aoe2DELanServer/models"
	"aoe2DELanServer/routes/game/party/shared"
	"net/http"
)

func PeerAdd(w http.ResponseWriter, r *http.Request) {
	adv, length, profileIds, raceIds, statGroupIds, teamIds := shared.ParseParameters(r)
	if adv == nil {
		i.JSON(&w, i.A{2})
		return
	}
	sess, _ := middleware.Session(r)
	currentUser := sess.GetUser()
	// Only the host can add peers
	host := adv.GetHost().GetUser()
	if host != currentUser {
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
		adv.NewPeer(u, raceIds[j], teamIds[j])
	}
	i.JSON(&w, i.A{0})
}
