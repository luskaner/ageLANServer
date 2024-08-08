package invitation

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/wss"
	"net/http"
)

type extendRequest struct {
	cancelRequest
	AdvertisementPassword string `schema:"gatheringpassword"`
}

func ExtendInvitation(w http.ResponseWriter, r *http.Request) {
	var q extendRequest
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	sess, _ := middleware.Session(r)
	u := sess.GetUser()
	adv, ok := models.GetAdvertisement(q.AdvertisementId)
	if !ok || adv.GetPasswordValue() != q.AdvertisementPassword {
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
	if invited {
		i.JSON(&w, i.A{0})
		return
	}
	peer.Invite(invitee)
	inviteeSession, ok := models.GetSessionByUser(invitee)
	if ok {
		go func() {
			// TODO: Wait for client to acknowledge it
			_ = wss.SendMessage(
				inviteeSession.GetId(),
				i.A{
					0,
					"ExtendInvitationMessage",
					q.UserId,
					i.A{
						u.GetProfileInfo(false),
						q.AdvertisementId,
						q.AdvertisementPassword,
					},
				},
			)
		}()
	} // TODO: If the user is offline send it when it comes online
	i.JSON(&w, i.A{0})
}
