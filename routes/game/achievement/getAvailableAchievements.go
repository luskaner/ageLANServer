package achievement

import (
	"aoe2DELanServer/files"
	"net/http"
)

func GetAvailableAchievements(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("achievements.json", &w, r, false)
}
