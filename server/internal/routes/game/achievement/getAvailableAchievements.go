package achievement

import (
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/models"
)

func GetAvailableAchievements(w http.ResponseWriter, r *http.Request) {
	models.G(r).Resources().ReturnSignedAsset("achievements.json", &w, r, false)
}
