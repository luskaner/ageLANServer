package leaderboard

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetLeaderBoard(c *gin.Context) {
	c.JSON(http.StatusOK, i.A{0, i.A{}, i.A{}, i.A{}})
}
