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
	game := models.G(r)
	u, _ := game.Users().GetUserById(sess.GetUserId())
	adv, ok := game.Advertisements().GetAdvertisement(q.AdvertisementId)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	var inviter *models.MainUser
	inviter, ok = game.Users().GetUserById(q.InviterId)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	var peer *models.MainPeer
	peer, ok = adv.GetPeer(inviter)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	if !peer.IsInvited(u) {
		i.JSON(&w, i.A{2})
		return
	}
	peer.Uninvite(u)
	var inviterSession *models.Session
	inviterSession, ok = models.GetSessionByUserId(inviter.GetId())
	if ok {
		var acceptStr string
		if q.Accept {
			acceptStr = "1"
		} else {
			acceptStr = "0"
		}
		wss.SendOrStoreMessage(
			inviterSession,
			"ReplyInvitationMessage",
			i.A{
				u.GetProfileInfo(false),
				q.AdvertisementId,
				acceptStr,
			},
		)
	}
	i.JSON(&w, i.A{0})
}
