package relationship

import (
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

func GetPresenceData(w http.ResponseWriter, r *http.Request) {
	models.G(r).Resources().ReturnSignedAsset("presenceData.json", &w, r, false)
}
