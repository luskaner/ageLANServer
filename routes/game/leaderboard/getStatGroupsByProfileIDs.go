package leaderboard

import (
	"aoe2DELanServer/routes/game/leaderboard/shared"
	"github.com/gin-gonic/gin"
)

func GetStatGroupsByProfileIDs(c *gin.Context) {
	response := shared.GetStatGroups(c.Query("profileids"), true, true)
	c.JSON(200, response)
}
