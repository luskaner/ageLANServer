package advertisement

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"

	"net/http"
)

func FindObservableAdvertisements(w http.ResponseWriter, r *http.Request) {
	var q wanQuery
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	game := models.G(r)
	title := game.Title()
	if title == common.GameAoE3 {
		observerGroupID := r.URL.Query().Get("observerGroupID")
		if observerGroupID != "0" {
			i.JSON(&w, i.A{2, i.A{}, i.A{}})
			return
		}
	}
	findAdvertisements(w, r, q.Length, q.Offset, true, nil, func(advertisement *models.MainAdvertisement) bool {
		return advertisement.UnsafeGetObserversEnabled()
	})
}
