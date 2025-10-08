package invitation

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
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
	game := models.G(r)
	advertisements := game.Advertisements()
	adv, ok := advertisements.GetAdvertisement(q.AdvertisementId)
	if ok {
		var password string
		advertisements.WithReadLock(adv.GetId(), func() {
			password = adv.UnsafeGetPasswordValue()
		})
		ok = password == q.AdvertisementPassword
	}
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	peers := adv.GetPeers()
	var peer *models.MainPeer
	sess := middleware.SessionOrPanic(r)
	u, ok := game.Users().GetUserById(sess.GetUserId())
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	peer, ok = peers.Load(u.GetId())
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	var invitee *models.MainUser
	invitee, ok = game.Users().GetUserById(q.UserId)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	if !peer.Invite(invitee) {
		i.JSON(&w, i.A{0})
		return
	}
	var inviteeSession *models.Session
	inviteeSession, ok = models.GetSessionByUserId(invitee.GetId())
	if ok {
		wss.SendOrStoreMessage(
			inviteeSession,
			"ExtendInvitationMessage",
			i.A{
				u.GetProfileInfo(false, game.Title(), inviteeSession.GetClientLibVersion()),
				q.AdvertisementId,
				q.AdvertisementPassword,
			},
		)
	}
	i.JSON(&w, i.A{0})
}
