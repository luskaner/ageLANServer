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
	game := models.G(r)
	u, _ := game.Users().GetUserById(sess.GetUserId())
	adv, ok := game.Advertisements().GetAdvertisement(q.AdvertisementId)
	if !ok {
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
	if !peer.IsInvited(invitee) {
		i.JSON(&w, i.A{0})
		return
	}
	peer.Uninvite(invitee)
	var inviteeSession *models.Session
	inviteeSession, ok = models.GetSessionByUserId(invitee.GetId())
	if ok {
		wss.SendOrStoreMessage(
			inviteeSession,
			"CancelInvitationMessage",
			i.A{
				u.GetProfileInfo(false),
				q.AdvertisementId,
			},
		)
	}
	i.JSON(&w, i.A{0})
}
