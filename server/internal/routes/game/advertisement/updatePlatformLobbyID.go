package advertisement

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
	"iter"
	"net/http"
)

type request struct {
	PlatformSessionId uint64 `schema:"platformlobbyID"`
	MatchID           int32  `schema:"matchID"`
}

func UpdatePlatformLobbyID(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := i.Bind(r, &req); err != nil {
		i.JSON(&w, i.A{2})
		return
	}

	game := models.G(r)
	advertisements := game.Advertisements()
	var currentUserId int32
	var peersId iter.Seq[int32]
	var ok bool
	advertisements.WithWriteLock(req.MatchID, func() {
		var adv *models.MainAdvertisement
		adv, ok = advertisements.GetAdvertisement(req.MatchID)
		if !ok {
			return
		}

		sess := middleware.Session(r)
		currentUserId = sess.GetUserId()
		peers := adv.GetPeers()
		if _, ok = peers.Load(currentUserId); !ok {
			return
		}

		adv.UnsafeUpdatePlatformSessionId(req.PlatformSessionId)

		_, peersId = peers.Keys()
		ok = true
	})
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	message := i.A{req.MatchID, "0", req.PlatformSessionId}
	for peerId := range peersId {
		if currentUserId == peerId {
			continue
		}
		if currentSess, ok := models.GetSessionByUserId(peerId); ok {
			wss.SendOrStoreMessage(
				currentSess,
				"PlatformSessionUpdateMessage",
				message,
			)
		}
	}

	i.JSON(&w,
		i.A{0},
	)
}
