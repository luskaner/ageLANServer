package relationship

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/session"
	"aoe2DELanServer/user"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetRelationships(c *gin.Context) {
	// As we don't have knowledge of friends, nor it is supposed to be many players on the server
	// just return all online users as if they were friends
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*session.Info)
	currentUser := sess.GetUser()
	profileInfo := user.GetProfileInfo(true, func(u *user.User) bool {
		return u != currentUser && u.GetPresence() > 0
	})
	c.JSON(http.StatusOK, j.A{0, j.A{}, j.A{}, j.A{}, j.A{}, profileInfo, j.A{}, j.A{}})
}
