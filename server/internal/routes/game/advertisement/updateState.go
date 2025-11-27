package advertisement

import (
	"iter"
	"net/http"
	"strconv"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	challengeShared "github.com/luskaner/ageLANServer/server/internal/routes/game/challenge/shared"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

type updateStateRequest struct {
	AdvertisementId int32 `schema:"advertisementid"`
	State           int8  `schema:"state"`
}

func UpdateState(w http.ResponseWriter, r *http.Request) {
	var q updateStateRequest
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	game := models.G(r)
	gameTitle := game.Title()
	advertisements := game.Advertisements()
	battleServers := game.BattleServers()
	var ok bool
	var peersLen int
	var peers iter.Seq2[int32, models.Peer]
	var advStartTime int64
	var advEncoded i.A
	advertisements.WithWriteLock(q.AdvertisementId, func() {
		var adv models.Advertisement
		adv, ok = game.Advertisements().GetAdvertisement(q.AdvertisementId)
		if !ok {
			i.JSON(&w, i.A{2})
			return
		}
		adv.UnsafeUpdateState(q.State)
		if adv.UnsafeGetState() == 1 {
			peersLen, peers = adv.GetPeers().Iter()
			advEncoded = adv.UnsafeEncode(gameTitle, battleServers)
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
			challengeProgress[j] = i.A{userIdSingleStr, challengeShared.GetChallengeProgressData()}
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
