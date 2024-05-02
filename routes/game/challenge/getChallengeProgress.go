package challenge

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/routes/game/challenge/shared"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetChallengeProgress(c *gin.Context) {
	c.JSON(http.StatusOK, i.A{0, shared.GetChallengeProgressData()})
}
