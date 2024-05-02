package advertisement

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"aoe2DELanServer/routes/game/advertisement/shared"
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
