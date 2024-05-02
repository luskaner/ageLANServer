package invitation

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"aoe2DELanServer/routes/game/invitation/shared"
	"aoe2DELanServer/routes/wss"
	"github.com/gin-gonic/gin"
	"net/http"
)

type cancelRequest struct {
	shared.Request
	UserId int32 `form:"inviteeid"`
}

func CancelInvitation(c *gin.Context) {
	var q cancelRequest
	if err := c.ShouldBind(&q); err != nil {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	sess, _ := c.Get("session")
	u := sess.(*models.Info).GetUser()
	adv, ok := models.GetAdvertisement(q.AdvertisementId)
	if !ok {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	peer, ok := adv.GetPeer(u)
	if !ok {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	invitee, ok := models.GetUserById(q.UserId)
	if !ok {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	invited := peer.IsInvited(invitee)
	if !invited {
		c.JSON(http.StatusOK, i.A{0})
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
	c.JSON(http.StatusOK, i.A{0})
}
