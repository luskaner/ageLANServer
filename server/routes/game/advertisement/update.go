package advertisement

import (
	"net/http"
	i "server/internal"
	"server/models"
	"server/routes/game/advertisement/shared"
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
