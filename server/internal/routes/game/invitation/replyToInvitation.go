package invitation

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/invitation/shared"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/wss"
	"net/http"
)

type replyRequest struct {
	shared.Request
	Accept    bool  `schema:"invitationreply"`
	InviterId int32 `schema:"inviterid"`
}

func ReplyToInvitation(w http.ResponseWriter, r *http.Request) {
	var q replyRequest
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
	inviter, ok := models.GetUserById(q.InviterId)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	peer, ok := adv.GetPeer(inviter)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	invited := peer.IsInvited(u)
	if !invited {
		i.JSON(&w, i.A{2})
		return
	}
	peer.Uninvite(u)
	inviterSession, ok := models.GetSessionByUser(inviter)
	if ok {
		var acceptStr string
		if q.Accept {
			acceptStr = "1"
		} else {
			acceptStr = "0"
		}
		go func() {
			// TODO: Wait for client to acknowledge it
			_ = wss.SendMessage(
				inviterSession.GetId(),
				i.A{
					0,
					"ReplyInvitationMessage",
					q.InviterId,
					i.A{
						u.GetProfileInfo(false),
						q.AdvertisementId,
						acceptStr,
					},
				},
			)
		}()
	} // TODO: If the user is offline send it when it comes online
	i.JSON(&w, i.A{0})
}
