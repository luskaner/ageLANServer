package leaderboard

import (
	"aoe2DELanServer/routes/game/leaderboard/extra"
	"github.com/gin-gonic/gin"
)

func GetPartyStat(c *gin.Context) {
	response := extra.GetStatGroups(c.Query("statsids"), false, true)
	c.JSON(200, response)
}
