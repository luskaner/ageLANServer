package achievement

import (
	"net/http"
	"server/files"
)

func GetAvailableAchievements(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("achievements.json", &w, r, false)
}
