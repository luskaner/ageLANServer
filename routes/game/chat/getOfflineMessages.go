package chat

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/session"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetOfflineMessages(c *gin.Context) {
	// TODO: What even are chat channels? plus the server seems to always return the same thing
	sessAny, _ := c.Get("session")
	sess := sessAny.(*session.Info)
	c.JSON(http.StatusOK, j.A{0, j.A{}, j.A{j.A{strconv.Itoa(int(sess.GetUser().GetId())), j.A{}}}, j.A{}, j.A{}, j.A{}})
}
