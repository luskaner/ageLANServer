package party

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
	"strconv"
)

func UpdateHost(w http.ResponseWriter, r *http.Request) {
	advStr := r.PostFormValue("match_id")
	advId, err := strconv.ParseInt(advStr, 10, 32)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	_, ok := models.G(r).Advertisements().GetAdvertisement(int32(advId))
	if !ok {
		i.JSON(&w, i.A{2})
	} else {
		i.JSON(&w, i.A{1})
	}
}
