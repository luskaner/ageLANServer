package leaderboard

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ApplyOfflineUpdates(c *gin.Context) {
	// TODO: Implement? which kind of updates?
	c.JSON(http.StatusOK, j.A{0, j.A{}, j.A{}})
}
