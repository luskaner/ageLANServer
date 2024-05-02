package leaderboard

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SetAvatarStatValues(c *gin.Context) {
	// TODO: Implement?
	c.JSON(http.StatusOK, i.A{0})
}
