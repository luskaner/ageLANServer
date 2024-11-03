package invitation

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
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
	game := models.G(r)
	u, _ := game.Users().GetUserById(sess.GetUserId())
	adv, ok := game.Advertisements().GetAdvertisement(q.AdvertisementId)
	if !ok || adv.GetPasswordValue() != q.AdvertisementPassword {
		i.JSON(&w, i.A{2})
		return
	}
	var peer *models.MainPeer
	peer, ok = adv.GetPeer(u)
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
	if peer.IsInvited(invitee) {
		i.JSON(&w, i.A{0})
		return
	}
	peer.Invite(invitee)
	var inviteeSession *models.Session
	inviteeSession, ok = models.GetSessionByUserId(invitee.GetId())
	if ok {
		wss.SendOrStoreMessage(
			inviteeSession,
			"ExtendInvitationMessage",
			i.A{
				u.GetProfileInfo(false),
				q.AdvertisementId,
				q.AdvertisementPassword,
			},
		)
	}
	i.JSON(&w, i.A{0})
}
