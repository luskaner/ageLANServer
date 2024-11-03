package achievement

import (
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

func GetAvailableAchievements(w http.ResponseWriter, r *http.Request) {
	models.G(r).Resources().ReturnSignedAsset("achievements.json", &w, r, false)
}
