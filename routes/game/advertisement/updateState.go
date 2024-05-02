package advertisement

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"aoe2DELanServer/routes/game/challenge/shared"
	"aoe2DELanServer/routes/wss"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func UpdateState(c *gin.Context) {
	stateStr := c.PostForm("state")
	state, err := strconv.ParseInt(stateStr, 10, 8)
	if err != nil {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	advStr := c.PostForm("advertisementid")
	advId, err := strconv.ParseUint(advStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	adv, ok := models.GetAdvertisement(uint32(advId))
	if !ok {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	previousState := adv.GetState()
	adv.UpdateState(int8(state))
	newState := adv.GetState()
	if previousState != newState && newState == 1 {
		peers := adv.GetPeers()
		peersLen := peers.Len()
		userIds := make([]i.A, peersLen)
		userIdStr := make([]i.A, peersLen)
		races := make([]i.A, peersLen)
		challengeProgress := make([]i.A, peersLen)
		sessionIds := make([]string, peersLen)
		j := 0
		for el := adv.GetPeers().Front(); el != nil; el = el.Next() {
			peer := el.Value
			sess, ok := models.GetSessionByUser(peer.GetUser())
			if !ok {
				continue
			}
			userIds[j] = i.A{peer.GetUser().GetId(), i.A{}}
			userIdStr[j] = i.A{strconv.Itoa(int(peer.GetUser().GetId())), i.A{}}
			races[j] = i.A{peer.GetUser().GetId(), strconv.Itoa(int(peer.GetRace()))}
			challengeProgress[j] = i.A{strconv.Itoa(int(peer.GetUser().GetId())), shared.GetChallengeProgressData()}
			sessionIds[j] = sess.GetId()
			j++
		}
		message := i.A{
			0,
			"MatchStartMessage",
			adv.GetHost().GetUser().GetId(),
			i.A{
				userIds,
				races,
				adv.GetStartTime(),
				userIdStr,
				adv.Encode(),
				challengeProgress,
			},
		}
		for _, sessionId := range sessionIds {
			go func() {
				_ = wss.SendMessage(sessionId, message)
			}()
		}
	}
	c.JSON(http.StatusOK, i.A{0})
}
