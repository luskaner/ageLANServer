package leaderboard

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetRecentMatchHistory(c *gin.Context) {
	// TODO: Implement? just in memory does not make much sense
	c.JSON(http.StatusOK, i.A{0, i.A{}})
}
