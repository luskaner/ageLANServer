package chat

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetOfflineMessages(c *gin.Context) {
	// TODO: What even are chat channels? plus the server seems to always return the same thing
	sessAny, _ := c.Get("session")
	sess := sessAny.(*models.Info)
	c.JSON(http.StatusOK, i.A{0, i.A{}, i.A{i.A{strconv.Itoa(int(sess.GetUser().GetId())), i.A{}}}, i.A{}, i.A{}, i.A{}})
}
