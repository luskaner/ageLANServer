package advertisement

import (
	"net/http"
	i "server/internal"
	"server/models"
	"server/routes/game/challenge/shared"
	"server/routes/wss"
	"strconv"
)

func UpdateState(w http.ResponseWriter, r *http.Request) {
	stateStr := r.PostFormValue("state")
	state, err := strconv.ParseInt(stateStr, 10, 8)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	advStr := r.PostFormValue("advertisementid")
	advId, err := strconv.ParseInt(advStr, 10, 32)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	adv, ok := models.GetAdvertisement(int32(advId))
	if !ok {
		i.JSON(&w, i.A{2})
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
		for el := adv.GetPeers().Oldest(); el != nil; el = el.Next() {
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
	i.JSON(&w, i.A{0})
}
