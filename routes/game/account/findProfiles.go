package account

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/session"
	"aoe2DELanServer/user"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func FindProfiles(c *gin.Context) {
	name := strings.ToLower(c.Query("name"))
	if len(name) < 1 {
		c.JSON(http.StatusOK, j.A{2, j.A{}})
		return
	}
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*session.Info)
	u := sess.GetUser()
	profileInfo := user.GetProfileInfo(true, func(currentUser *user.User) bool {
		if u == currentUser {
			return false
		}
		return strings.Contains(strings.ToLower(currentUser.GetAlias()), name)
	})
	c.JSON(http.StatusOK, j.A{0, profileInfo})
}
