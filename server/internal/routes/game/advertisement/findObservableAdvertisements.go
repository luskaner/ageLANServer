package advertisement

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"

	"net/http"
)

func findObsAdvErr(w *http.ResponseWriter) {
	i.JSON(w, i.A{2, i.A{}, i.A{}})
}

func FindObservableAdvertisements(w http.ResponseWriter, r *http.Request) {
	var q wanQuery
	if err := i.Bind(r, &q); err != nil {
		findObsAdvErr(&w)
		return
	}
	game := models.G(r)
	title := game.Title()
	if title == common.GameAoE3 {
		observerGroupID := r.URL.Query().Get("observerGroupID")
		if observerGroupID != "0" {
			findObsAdvErr(&w)
			return
		}
	}
	findAdvertisements(w, r, q.Length, q.Offset, true, nil, func(advertisement models.Advertisement) bool {
		return advertisement.UnsafeGetObserversEnabled()
	})
}
