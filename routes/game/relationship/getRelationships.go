package relationship

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetRelationships(c *gin.Context) {
	// As we don't have knowledge of friends, nor it is supposed to be many players on the server
	// just return all online users as if they were friends
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*models.Info)
	currentUser := sess.GetUser()
	profileInfo := models.GetProfileInfo(true, func(u *models.User) bool {
		return u != currentUser && u.GetPresence() > 0
	})
	c.JSON(http.StatusOK, i.A{0, i.A{}, i.A{}, i.A{}, i.A{}, profileInfo, i.A{}, i.A{}})
}
