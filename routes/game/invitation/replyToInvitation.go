package invitation

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"aoe2DELanServer/routes/game/invitation/shared"
	"aoe2DELanServer/routes/wss"
	"github.com/gin-gonic/gin"
	"net/http"
)

type replyRequest struct {
	shared.Request
	Accept    bool  `form:"invitationreply"`
	InviterId int32 `form:"inviterid"`
}

func ReplyToInvitation(c *gin.Context) {
	var q replyRequest
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
	inviter, ok := models.GetUserById(q.InviterId)
	if !ok {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	peer, ok := adv.GetPeer(inviter)
	if !ok {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	invited := peer.IsInvited(u)
	if !invited {
		c.JSON(http.StatusOK, i.A{2})
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
	c.JSON(http.StatusOK, i.A{0})
}
