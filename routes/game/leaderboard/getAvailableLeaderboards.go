package leaderboard

import (
	"aoe2DELanServer/files"
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAvailableLeaderboards(c *gin.Context) {
	file := files.Config["leaderboards.json"]
	response := make(i.A, len(file))
	copy(response, file)
	response = append(i.A{0}, response...)
	c.JSON(http.StatusOK, response)
}
