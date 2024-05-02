package challenge

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/routes/game/challenge/extra"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetChallengeProgress(c *gin.Context) {
	c.JSON(http.StatusOK, j.A{0, extra.GetChallengeProgressData()})
}
