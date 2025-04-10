package advertisement

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/challenge/shared"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
	"iter"
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
	advId64, err := strconv.ParseInt(advStr, 10, 32)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	game := models.G(r)
	gameTitle := game.Title()
	advertisements := game.Advertisements()
	advId := int32(advId64)
	var ok bool
	var peersLen int
	var peers iter.Seq2[int32, *models.MainPeer]
	var advStartTime int64
	var advEncoded i.A
	advertisements.WithWriteLock(advId, func() {
		var adv *models.MainAdvertisement
		adv, ok = game.Advertisements().GetAdvertisement(advId)
		if !ok {
			i.JSON(&w, i.A{2})
			return
		}
		adv.UnsafeUpdateState(int8(state))
		if adv.UnsafeGetState() == 1 {
			peersLen, peers = adv.GetPeers().Iter()
			advEncoded = adv.UnsafeEncode(gameTitle)
			advStartTime = adv.UnsafeGetStartTime()
		}
		ok = true
	})
	if ok {
		userIds := make([]i.A, peersLen)
		userIdStr := make([]i.A, peersLen)
		races := make([]i.A, peersLen)
		challengeProgress := make([]i.A, peersLen)
		sessions := make([]*models.Session, peersLen)
		j := 0
		for userId, peer := range peers {
			var sess *models.Session
			sess, ok = models.GetSessionByUserId(userId)
			if !ok {
				continue
			}
			userIdSingleStr := strconv.Itoa(int(userId))
			userIds[j] = i.A{userId, i.A{}}
			userIdStr[j] = i.A{userIdSingleStr, i.A{}}
			peerMutable := peer.GetMutable()
			races[j] = i.A{userIdSingleStr, strconv.Itoa(int(peerMutable.Race))}
			challengeProgress[j] = i.A{userIdSingleStr, shared.GetChallengeProgressData()}
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
					advStartTime,
					userIdStr,
					advEncoded,
					challengeProgress,
				},
			)
		}
		i.JSON(&w, i.A{0})
	} else {
		i.JSON(&w, i.A{2})
	}
}
