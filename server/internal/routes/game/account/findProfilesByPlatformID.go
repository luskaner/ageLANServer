package account

import (
	"encoding/json"
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func FindProfilesByPlatformID(w http.ResponseWriter, r *http.Request) {
	platformIdsStr := r.PostFormValue("platformIDs")
	if len(platformIdsStr) < 1 {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	var platformIds []uint64
	err := json.Unmarshal([]byte(platformIdsStr), &platformIds)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	platformIdsMap := make(map[uint64]interface{}, len(platformIds))
	for _, platformId := range platformIds {
		platformIdsMap[platformId] = struct{}{}
	}
	game := models.G(r)
	gameTitle := game.Title()
	sess := models.SessionOrPanic(r)
	profileInfo := game.Users().GetProfileInfo(true, func(currentUser *models.MainUser) bool {
		_, ok := platformIdsMap[currentUser.GetPlatformUserID()]
		return ok
	}, gameTitle, sess.GetClientLibVersion())
	i.JSON(&w, i.A{0, profileInfo})
}
