package party

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

type advRequest struct {
	MatchID int32 `schema:"match_id"`
}

func UpdateHost(w http.ResponseWriter, r *http.Request) {
	var req advRequest
	err := i.Bind(r, &req)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	_, ok := models.G(r).Advertisements().GetAdvertisement(req.MatchID)
	if !ok {
		i.JSON(&w, i.A{2})
	} else {
		i.JSON(&w, i.A{1})
	}
}
