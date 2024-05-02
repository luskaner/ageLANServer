package leaderboard

import (
	"aoe2DELanServer/routes/game/leaderboard/shared"
	"github.com/gin-gonic/gin"
)

func GetPartyStat(c *gin.Context) {
	response := shared.GetStatGroups(c.Query("statsids"), false, true)
	c.JSON(200, response)
}
