package relationship

import (
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/models"
)

func GetPresenceData(w http.ResponseWriter, r *http.Request) {
	models.G(r).Resources().ReturnSignedAsset("presenceData.json", &w, r, false)
}
