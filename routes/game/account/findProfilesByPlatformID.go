package account

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/session"
	"aoe2DELanServer/user"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func FindProfilesByPlatformID(c *gin.Context) {
	platformIdsStr := c.PostForm("platformIDs")
	if len(platformIdsStr) < 1 {
		c.JSON(http.StatusOK, j.A{2, j.A{}})
		return
	}
	var platformIds []uint64
	err := json.Unmarshal([]byte(platformIdsStr), &platformIds)
	if err != nil {
		c.JSON(http.StatusOK, j.A{2, j.A{}})
		return
	}
	platformIdsMap := make(map[uint64]interface{}, len(platformIds))
	for _, platformId := range platformIds {
		platformIdsMap[platformId] = struct{}{}
	}
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*session.Info)
	u := sess.GetUser()
	profileInfo := user.GetProfileInfo(true, func(currentUser *user.User) bool {
		if u == currentUser {
			return false
		}
		_, ok := platformIdsMap[currentUser.GetPlatformUserID()]
		return ok
	})
	c.JSON(http.StatusOK, j.A{0, profileInfo})
}
