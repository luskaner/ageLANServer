package party

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/party/shared"
)

func PeerAdd(w http.ResponseWriter, r *http.Request) {
	parseError, advId, length, profileIds, raceIds, teamIds := shared.ParseParameters(r)
	if parseError {
		i.JSON(&w, i.A{2})
		return
	}
	sess := middleware.SessionOrPanic(r)
	currentUserId := sess.GetUserId()
	game := models.G(r)
	gameUsers := game.Users()
	advertisements := game.Advertisements()
	var ok bool
	advertisements.WithWriteLock(advId, func() {
		adv, exists := advertisements.GetAdvertisement(advId)
		if !exists {
			return
		}
		// Only the host can add peers
		if hostId := adv.GetHostId(); hostId != currentUserId {
			return
		}
		users := make([]*models.MainUser, length)
		for j := 0; j < length; j++ {
			var u *models.MainUser
			u, ok = gameUsers.GetUserById(profileIds[j])
			if !ok {
				return
			}
			users[j] = u
		}
		advIp := adv.GetIp()
		var addedUserIds []int32
		for j, u := range users {
			if peer := advertisements.UnsafeNewPeer(advId, advIp, u.GetId(), u.GetStatId(), raceIds[j], teamIds[j]); peer != nil {
				addedUserIds = append(addedUserIds, u.GetId())
			} else {
				break
			}
		}
		if len(addedUserIds) == len(users) {
			ok = true
		} else {
			for _, userId := range addedUserIds {
				advertisements.UnsafeRemovePeer(advId, userId)
			}
		}
	})
	if ok {
		i.JSON(&w, i.A{0})
	} else {
		i.JSON(&w, i.A{2})
	}
}
