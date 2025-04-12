package account

import (
	"encoding/json"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

func GetProfileName(w http.ResponseWriter, r *http.Request) {
	profileIdsStr := r.URL.Query().Get("profile_ids")
	if len(profileIdsStr) < 1 {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	var profileIds []int32
	err := json.Unmarshal([]byte(profileIdsStr), &profileIds)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	profileIdsMap := make(map[int32]interface{}, len(profileIds))
	for _, platformId := range profileIds {
		profileIdsMap[platformId] = struct{}{}
	}
	game := models.G(r)
	gameTitle := game.Title()
	sess := middleware.Session(r)
	profileInfo := game.Users().GetProfileInfo(false, func(currentUser *models.MainUser) bool {
		_, ok := profileIdsMap[currentUser.GetId()]
		return ok
	}, gameTitle, sess.GetClientLibVersion())
	i.JSON(&w, i.A{0, profileInfo})
}
