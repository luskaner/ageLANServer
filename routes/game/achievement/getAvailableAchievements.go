package achievement

import (
	"aoe2DELanServer/files"
	"github.com/gin-gonic/gin"
)

func GetAvailableAchievements(c *gin.Context) {
	files.ReturnSignedAsset("achievements.json", c, false)
}
