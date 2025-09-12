package advertisement

import (
	"iter"
	"net/http"
	"strconv"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

type request struct {
	MatchID int32 `schema:"matchID"`
}

func updatePlatformID(w *http.ResponseWriter, r *http.Request, idKey string) {
	var req request
	if err := i.Bind(r, &req); err != nil {
		i.JSON(w, i.A{2})
		return
	}

	game := models.G(r)
	advertisements := game.Advertisements()
	var currentUserId int32
	var peersId iter.Seq[int32]
	var ok bool
	var metadata string
	idValue, err := strconv.ParseInt(r.PostFormValue(idKey), 10, 64)
	if err != nil {
		i.JSON(w, i.A{2})
		return
	}
	idValueUint := uint64(idValue)
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

		adv.UnsafeUpdatePlatformSessionId(idValueUint)
		metadata = adv.GetMetadata()
		_, peersId = peers.Keys()
		ok = true
	})
	if !ok {
		i.JSON(w, i.A{2})
		return
	}
	message := i.A{req.MatchID, metadata, idValueUint}
	for peerId := range peersId {
		/*if gameTitle != common.GameAoE1 && currentUserId == peerId {
			continue
		}*/
		if currentSess, ok := models.GetSessionByUserId(peerId); ok {
			wss.SendOrStoreMessage(
				currentSess,
				"PlatformSessionUpdateMessage",
				message,
			)
		}
	}

	i.JSON(w,
		i.A{0},
	)
}

func UpdatePlatformLobbyID(w http.ResponseWriter, r *http.Request) {
	updatePlatformID(&w, r, "platformlobbyID")
}
