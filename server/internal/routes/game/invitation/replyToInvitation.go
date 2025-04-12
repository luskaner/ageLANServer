package invitation

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/invitation/shared"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
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
	game := models.G(r)
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
	peers := adv.GetPeers()
	var peer *models.MainPeer
	sess := middleware.Session(r)
	u, ok := game.Users().GetUserById(sess.GetUserId())
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	peer, ok = peers.Load(inviter.GetId())
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	if !peer.Uninvite(u) {
		i.JSON(&w, i.A{0})
		return
	}
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
				u.GetProfileInfo(false, game.Title(), inviterSession.GetClientLibVersion()),
				q.AdvertisementId,
				acceptStr,
			},
		)
	}
	i.JSON(&w, i.A{0})
}
