package invitation

import (
	"net/http"
	i "server/internal"
	"server/middleware"
	"server/models"
	"server/routes/game/invitation/shared"
	"server/routes/wss"
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
	peer, ok := adv.GetPeer(u)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	invitee, ok := models.GetUserById(q.UserId)
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
	inviteeSession, ok := models.GetSessionByUser(invitee)
	if ok {
		go func() {
			// TODO: Wait for client to acknowledge it
			_ = wss.SendMessage(
				inviteeSession.GetId(),
				i.A{
					0,
					"CancelInvitationMessage",
					q.UserId,
					i.A{
						u.GetProfileInfo(false),
						q.AdvertisementId,
					},
				},
			)
		}()
	} // TODO: If the user is offline send it when it comes online
	i.JSON(&w, i.A{0})
}
