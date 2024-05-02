package leaderboard

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetRecentMatchHistory(c *gin.Context) {
	// TODO: Implement? just in memory does not make much sense
	c.JSON(http.StatusOK, j.A{0, j.A{}})
}
