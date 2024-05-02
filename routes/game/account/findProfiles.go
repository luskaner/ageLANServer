package account

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func FindProfiles(c *gin.Context) {
	name := strings.ToLower(c.Query("name"))
	if len(name) < 1 {
		c.JSON(http.StatusOK, i.A{2, i.A{}})
		return
	}
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*models.Info)
	u := sess.GetUser()
	profileInfo := models.GetProfileInfo(true, func(currentUser *models.User) bool {
		if u == currentUser {
			return false
		}
		return strings.Contains(strings.ToLower(currentUser.GetAlias()), name)
	})
	c.JSON(http.StatusOK, i.A{0, profileInfo})
}
