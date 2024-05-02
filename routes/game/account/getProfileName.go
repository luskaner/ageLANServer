package account

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetProfileName(c *gin.Context) {
	profileIdsStr := c.Query("profile_ids")
	if len(profileIdsStr) < 1 {
		c.JSON(http.StatusOK, i.A{2, i.A{}})
		return
	}
	var profileIds []int32
	err := json.Unmarshal([]byte(profileIdsStr), &profileIds)
	if err != nil {
		c.JSON(http.StatusOK, i.A{2, i.A{}})
		return
	}
	profileIdsMap := make(map[int32]interface{}, len(profileIds))
	for _, platformId := range profileIds {
		profileIdsMap[platformId] = struct{}{}
	}
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*models.Info)
	u := sess.GetUser()
	profileInfo := models.GetProfileInfo(false, func(currentUser *models.User) bool {
		if u == currentUser {
			return false
		}
		_, ok := profileIdsMap[currentUser.GetId()]
		return ok
	})
	c.JSON(http.StatusOK, i.A{0, profileInfo})
}
