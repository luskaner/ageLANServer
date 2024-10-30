package account

import (
	"encoding/json"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"net/http"
)

func FindProfilesByPlatformID(w http.ResponseWriter, r *http.Request) {
	// TODO: Check it works in AoE3 if not empty
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
	sess, _ := middleware.Session(r)
	users := models.G(r).Users()
	u, _ := users.GetUserById(sess.GetUserId())
	profileInfo := users.GetProfileInfo(true, func(currentUser *models.MainUser) bool {
		if u == currentUser {
			return false
		}
		_, ok := platformIdsMap[currentUser.GetPlatformUserID()]
		return ok
	})
	i.JSON(&w, i.A{0, profileInfo})
}
