package relationship

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/session"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func SetPresence(c *gin.Context) {
	presenceId, ok := c.GetPostForm("presence_id")
	if !ok {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	presence, err := strconv.Atoi(presenceId)
	if err != nil {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*session.Info)
	sess.GetUser().SetPresence(int8(presence))
	c.JSON(http.StatusOK, j.A{0})
}
