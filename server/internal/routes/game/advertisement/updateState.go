package advertisement

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/challenge/shared"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
	"net/http"
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
	game := models.G(r)
	gameTitle := game.Title()
	adv, ok := game.Advertisements().GetAdvertisement(int32(advId))
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	adv.UpdateState(int8(state))
	if adv.GetState() == 1 {
		peers := adv.GetPeers()
		peersLen := peers.Len()
		userIds := make([]i.A, peersLen)
		userIdStr := make([]i.A, peersLen)
		races := make([]i.A, peersLen)
		challengeProgress := make([]i.A, peersLen)
		sessions := make([]*models.Session, peersLen)
		advEncoded := adv.Encode(gameTitle)
		j := 0
		for el := adv.GetPeers().Oldest(); el != nil; el = el.Next() {
			peer := el.Value
			var sess *models.Session
			sess, ok = models.GetSessionByUserId(peer.GetUser().GetId())
			if !ok {
				continue
			}
			userIds[j] = i.A{peer.GetUser().GetId(), i.A{}}
			userIdStr[j] = i.A{strconv.Itoa(int(peer.GetUser().GetId())), i.A{}}
			races[j] = i.A{peer.GetUser().GetId(), strconv.Itoa(int(peer.GetRace()))}
			challengeProgress[j] = i.A{strconv.Itoa(int(peer.GetUser().GetId())), shared.GetChallengeProgressData()}
			sessions[j] = sess
			j++
		}
		for _, session := range sessions {
			wss.SendOrStoreMessage(
				session,
				"MatchStartMessage",
				i.A{
					userIds,
					races,
					adv.GetStartTime(),
					userIdStr,
					advEncoded,
					challengeProgress,
				},
			)
		}
	}
	i.JSON(&w, i.A{0})
}
