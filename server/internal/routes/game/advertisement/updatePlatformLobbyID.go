package advertisement

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
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
	adv, ok := game.Advertisements().GetAdvertisement(req.MatchID)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}

	sess, _ := middleware.Session(r)
	u, _ := game.Users().GetUserById(sess.GetUserId())
	var peer *models.MainPeer
	if peer, ok = adv.GetPeer(u); !ok {
		i.JSON(&w, i.A{2})
		return
	}

	adv.UpdatePlatformSessionId(req.PlatformSessionId)
	message := i.A{req.MatchID, "0", req.PlatformSessionId}

	for el := adv.GetPeers().Oldest(); el != nil; el = el.Next() {
		if el.Value == peer {
			continue
		}
		if currentSess, ok := models.GetSessionByUserId(el.Value.GetUser().GetId()); ok {
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
