package party

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"aoe2DELanServer/routes/wss"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

type request struct {
	ToProfileIdsStr string `form:"to_profile_ids" binding:"required"`
	MessageTypeID   uint8  `form:"messageTypeID"`
	MatchID         uint32 `form:"match_id" binding:"required"`
	Broadcast       bool   `form:"broadcast" binding:"required"`
	Message         string `form:"message" binding:"required"`
}

func SendMatchChat(c *gin.Context) {
	var req request
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, i.A{2})
		return
	}

	var toProfileIds []int32
	err := json.Unmarshal([]byte(req.ToProfileIdsStr), &toProfileIds)
	if err != nil {
		c.JSON(http.StatusOK, i.A{2})
		return
	}

	adv, ok := models.GetAdvertisement(req.MatchID)
	if !ok {
		c.JSON(http.StatusOK, i.A{2})
		return
	}

	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*models.Info)
	currentUser := sess.GetUser()

	// Only peers within the match can send messages
	// TODO: What about AI?
	if _, ok := adv.GetPeer(currentUser); !ok {
		c.JSON(http.StatusOK, i.A{2})
		return
	}

	receivers := make([]*models.User, len(toProfileIds))
	for i, profileId := range toProfileIds {
		receivers[i], _ = models.GetUserById(profileId)
	}

	message := adv.AddMessage(
		req.Broadcast,
		req.Message,
		req.MessageTypeID,
		currentUser,
		receivers,
	)

	messageEncoded := message.Encode()
	for _, receiver := range receivers {
		receiverSession, ok := models.GetSessionByUser(receiver)
		if !ok {
			continue
		}
		go func() {
			wss.SendMessage(
				receiverSession.GetId(),
				i.A{
					0,
					"MatchReceivedChatMessage",
					receiver.GetId(),
					messageEncoded,
				},
			)
		}()
	}
	c.JSON(http.StatusOK, i.A{0})
}
