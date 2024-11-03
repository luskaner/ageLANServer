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
	sess, _ := middleware.Session(r)
	users := models.G(r).Users()
	u, _ := users.GetUserById(sess.GetUserId())
	profileInfo := users.GetProfileInfo(false, func(currentUser *models.MainUser) bool {
		if u == currentUser {
			return false
		}
		_, ok := profileIdsMap[currentUser.GetId()]
		return ok
	})
	i.JSON(&w, i.A{0, profileInfo})
}
