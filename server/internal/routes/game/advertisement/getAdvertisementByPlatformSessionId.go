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
	advertisements := models.G(r).Advertisements()
	adv := advertisements.UnsafeFirstAdvertisement(func(adv models.Advertisement) bool {
		var platform string
		var id uint64
		advertisements.WithReadLock(adv.GetId(), func() {
			platform, id = adv.UnsafeGetPlatformSessionId()
		},
		)
		return platform == req.Platform && req.SessionID == id
	})
	if adv == nil {
		i.JSON(&w, i.A{0, i.A{}})
	} else {
		i.JSON(&w, i.A{0, adv})
	}
}
