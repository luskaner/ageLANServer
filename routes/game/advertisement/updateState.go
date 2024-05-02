package advertisement

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/routes/game/advertisement/extra"
	extra2 "aoe2DELanServer/routes/game/challenge/extra"
	"aoe2DELanServer/routes/wss"
	"aoe2DELanServer/session"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func UpdateState(c *gin.Context) {
	stateStr := c.PostForm("state")
	state, err := strconv.ParseInt(stateStr, 10, 8)
	if err != nil {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	advStr := c.PostForm("advertisementid")
	advId, err := strconv.ParseUint(advStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	adv, ok := extra.Get(uint32(advId))
	if !ok {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	previousState := adv.GetState()
	adv.UpdateState(int8(state))
	newState := adv.GetState()
	if previousState != newState && newState == 1 {
		peers := adv.GetPeers()
		peersLen := peers.Len()
		userIds := make([]j.A, peersLen)
		userIdStr := make([]j.A, peersLen)
		races := make([]j.A, peersLen)
		challengeProgress := make([]j.A, peersLen)
		sessionIds := make([]string, peersLen)
		i := 0
		for el := adv.GetPeers().Front(); el != nil; el = el.Next() {
			peer := el.Value
			sess, ok := session.GetByUser(peer.GetUser())
			if !ok {
				continue
			}
			userIds[i] = j.A{peer.GetUser().GetId(), j.A{}}
			userIdStr[i] = j.A{strconv.Itoa(int(peer.GetUser().GetId())), j.A{}}
			races[i] = j.A{peer.GetUser().GetId(), strconv.Itoa(int(peer.GetRace()))}
			challengeProgress[i] = j.A{strconv.Itoa(int(peer.GetUser().GetId())), extra2.GetChallengeProgressData()}
			sessionIds[i] = sess.GetId()
			i++
		}
		message := j.A{
			0,
			"MatchStartMessage",
			adv.GetHost().GetUser().GetId(),
			j.A{
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
	c.JSON(http.StatusOK, j.A{0})
}
