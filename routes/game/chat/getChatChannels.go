package chat

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetChatChannels(c *gin.Context) {
	// TODO: What even are chat channels? plus the server seems to always return the same thing
	c.JSON(http.StatusOK, i.A{0, i.A{}, 100})
}
