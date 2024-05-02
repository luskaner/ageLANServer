package account

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SetCrossplayEnabled(c *gin.Context) {
	// Crossplay is always enabled regardless of the value sent
	enable := c.PostForm("enable")
	if enable == "1" {
		c.JSON(http.StatusOK, i.A{0})
	} else {
		// Do not accept disabling it
		c.JSON(http.StatusOK, i.A{2})
	}
}
