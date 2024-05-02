package account

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/session"
	"aoe2DELanServer/user"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetProfileName(c *gin.Context) {
	profileIdsStr := c.Query("profile_ids")
	if len(profileIdsStr) < 1 {
		c.JSON(http.StatusOK, j.A{2, j.A{}})
		return
	}
	var profileIds []int32
	err := json.Unmarshal([]byte(profileIdsStr), &profileIds)
	if err != nil {
		c.JSON(http.StatusOK, j.A{2, j.A{}})
		return
	}
	profileIdsMap := make(map[int32]interface{}, len(profileIds))
	for _, platformId := range profileIds {
		profileIdsMap[platformId] = struct{}{}
	}
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*session.Info)
	u := sess.GetUser()
	profileInfo := user.GetProfileInfo(false, func(currentUser *user.User) bool {
		if u == currentUser {
			return false
		}
		_, ok := profileIdsMap[currentUser.GetId()]
		return ok
	})
	c.JSON(http.StatusOK, j.A{0, profileInfo})
}
