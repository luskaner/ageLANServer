package leaderboard

import (
	"aoe2DELanServer/routes/game/leaderboard/extra"
	"github.com/gin-gonic/gin"
)

func GetStatGroupsByProfileIDs(c *gin.Context) {
	response := extra.GetStatGroups(c.Query("profileids"), true, true)
	c.JSON(200, response)
}
