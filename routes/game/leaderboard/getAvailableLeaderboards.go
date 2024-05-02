package leaderboard

import (
	"aoe2DELanServer/asset"
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAvailableLeaderboards(c *gin.Context) {
	file := asset.Files["leaderboards.json"]
	response := make(j.A, len(file))
	copy(response, file)
	response = append(j.A{0}, response...)
	c.JSON(http.StatusOK, response)
}
