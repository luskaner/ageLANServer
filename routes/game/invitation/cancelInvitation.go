package invitation

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/routes/game/advertisement/extra"
	invExtra "aoe2DELanServer/routes/game/invitation/extra"
	"aoe2DELanServer/routes/wss"
	"aoe2DELanServer/session"
	"aoe2DELanServer/user"
	"github.com/gin-gonic/gin"
	"net/http"
)

type cancelRequest struct {
	invExtra.Request
	UserId int32 `form:"inviteeid"`
}

func CancelInvitation(c *gin.Context) {
	var q cancelRequest
	if err := c.ShouldBind(&q); err != nil {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	sess, _ := c.Get("session")
	u := sess.(*session.Info).GetUser()
	adv, ok := extra.Get(q.AdvertisementId)
	if !ok {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	peer, ok := adv.GetPeer(u)
	if !ok {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	invitee, ok := user.GetById(q.UserId)
	if !ok {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	invited := peer.IsInvited(invitee)
	if !invited {
		c.JSON(http.StatusOK, j.A{0})
		return
	}
	peer.Uninvite(invitee)
	inviteeSession, ok := session.GetByUser(invitee)
	if ok {
		go func() {
			// TODO: Wait for client to acknowledge it
			_ = wss.SendMessage(
				inviteeSession.GetId(),
				j.A{
					0,
					"CancelInvitationMessage",
					q.UserId,
					j.A{
						u.GetProfileInfo(false),
						q.AdvertisementId,
					},
				},
			)
		}()
	} // TODO: If the user is offline send it when it comes online
	c.JSON(http.StatusOK, j.A{0})
}
