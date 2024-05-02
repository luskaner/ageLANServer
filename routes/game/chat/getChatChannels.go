package chat

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetChatChannels(c *gin.Context) {
	// TODO: What even are chat channels? plus the server seems to always return the same thing
	c.JSON(http.StatusOK, j.A{0, j.A{}, 100})
}
