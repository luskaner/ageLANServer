package leaderboard

import (
	"aoe2DELanServer/routes/game/leaderboard/extra"
	"github.com/gin-gonic/gin"
)

func GetStatsForLeaderboardByProfileName(c *gin.Context) {
	response := extra.GetStatGroups(c.Query("profileids"), true, false)
	c.JSON(200, response)
}
