package advertisement

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/models"
	"github.com/luskaner/aoe2DELanServer/server/routes/game/advertisement/shared"
	"net/http"
)

func Update(w http.ResponseWriter, r *http.Request) {
	var q shared.AdvertisementUpdateRequest
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	advOriginal, ok := models.GetAdvertisement(q.Id)
	if !ok {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	advOriginal.Update(&q)
	i.JSON(&w,
		i.A{
			0,
			advOriginal.Encode(),
		},
	)
}
