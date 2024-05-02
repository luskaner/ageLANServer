package leaderboard

import (
	"aoe2DELanServer/routes/game/leaderboard/shared"
	"github.com/gin-gonic/gin"
)

func GetStatsForLeaderboardByProfileName(c *gin.Context) {
	response := shared.GetStatGroups(c.Query("profileids"), true, false)
	c.JSON(200, response)
}
