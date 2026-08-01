package advertisement

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

type onlinePlatform struct {
	Platform string `schema:"onlinePlatform"`
}

type getAdvertisementByPlatformSessionIdRequest struct {
	onlinePlatform
	SessionID uint64 `schema:"platformSessionID"`
}

func GetAdvertisementByPlatformSessionId(w http.ResponseWriter, r *http.Request) {
	var req getAdvertisementByPlatformSessionIdRequest
	err := i.Bind(r, &req)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	game := models.G(r)
	advertisements := game.Advertisements()
	var advEncoded i.A
	_ = advertisements.UnsafeFirstAdvertisement(func(adv models.Advertisement) bool {
		var ok bool
		advertisements.WithReadLock(adv.GetId(), func() {
			platform, id := adv.UnsafeGetPlatformSessionId()
			if platform == req.Platform && req.SessionID == id {
				advEncoded = adv.UnsafeEncode(game.Title(), game.BattleServers())
				ok = true
			}
		},
		)
		return ok
	})
	if advEncoded == nil {
		i.JSON(&w, i.A{0, i.A{}})
	} else {
		i.JSON(&w, i.A{0, advEncoded})
	}
}
