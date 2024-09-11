package invitation

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/invitation/shared"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/wss"
	"net/http"
)

type cancelRequest struct {
	shared.Request
	UserId int32 `schema:"inviteeid"`
}

func CancelInvitation(w http.ResponseWriter, r *http.Request) {
	var q cancelRequest
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	sess, _ := middleware.Session(r)
	u := sess.GetUser()
	adv, ok := models.GetAdvertisement(q.AdvertisementId)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	var peer *models.Peer
	peer, ok = adv.GetPeer(u)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	var invitee *models.User
	invitee, ok = models.GetUserById(q.UserId)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	invited := peer.IsInvited(invitee)
	if !invited {
		i.JSON(&w, i.A{0})
		return
	}
	peer.Uninvite(invitee)
	var inviteeSession *models.Session
	inviteeSession, ok = models.GetSessionByUser(invitee)
	if ok {
		go func(userId int32, advertisementId int32, sessId string, userProfileInfo i.A) {
			_ = wss.SendMessage(
				sessId,
				i.A{
					0,
					"CancelInvitationMessage",
					userId,
					i.A{
						userProfileInfo,
						advertisementId,
					},
				},
			)
		}(q.UserId, q.AdvertisementId, inviteeSession.GetId(), u.GetProfileInfo(false))
	} // TODO: If the user is offline send it when it comes online?
	i.JSON(&w, i.A{0})
}
