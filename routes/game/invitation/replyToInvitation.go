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

type replyRequest struct {
	invExtra.Request
	Accept    bool  `form:"invitationreply"`
	InviterId int32 `form:"inviterid"`
}

func ReplyToInvitation(c *gin.Context) {
	var q replyRequest
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
	inviter, ok := user.GetById(q.InviterId)
	if !ok {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	peer, ok := adv.GetPeer(inviter)
	if !ok {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	invited := peer.IsInvited(u)
	if !invited {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	peer.Uninvite(u)
	inviterSession, ok := session.GetByUser(inviter)
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
				j.A{
					0,
					"ReplyInvitationMessage",
					q.InviterId,
					j.A{
						u.GetProfileInfo(false),
						q.AdvertisementId,
						acceptStr,
					},
				},
			)
		}()
	} // TODO: If the user is offline send it when it comes online
	c.JSON(http.StatusOK, j.A{0})
}
