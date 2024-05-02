package account

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func FindProfilesByPlatformID(c *gin.Context) {
	platformIdsStr := c.PostForm("platformIDs")
	if len(platformIdsStr) < 1 {
		c.JSON(http.StatusOK, i.A{2, i.A{}})
		return
	}
	var platformIds []uint64
	err := json.Unmarshal([]byte(platformIdsStr), &platformIds)
	if err != nil {
		c.JSON(http.StatusOK, i.A{2, i.A{}})
		return
	}
	platformIdsMap := make(map[uint64]interface{}, len(platformIds))
	for _, platformId := range platformIds {
		platformIdsMap[platformId] = struct{}{}
	}
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*models.Info)
	u := sess.GetUser()
	profileInfo := models.GetProfileInfo(true, func(currentUser *models.User) bool {
		if u == currentUser {
			return false
		}
		_, ok := platformIdsMap[currentUser.GetPlatformUserID()]
		return ok
	})
	c.JSON(http.StatusOK, i.A{0, profileInfo})
}
