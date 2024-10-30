package relationship

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"net/http"
)

func GetPresenceData(w http.ResponseWriter, r *http.Request) {
	models.G(r).Resources().ReturnSignedAsset("presenceData.json", &w, r, false)
}
