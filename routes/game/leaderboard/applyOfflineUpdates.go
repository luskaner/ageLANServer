package leaderboard

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ApplyOfflineUpdates(c *gin.Context) {
	// TODO: Implement? which kind of updates?
	c.JSON(http.StatusOK, i.A{0, i.A{}, i.A{}})
}
