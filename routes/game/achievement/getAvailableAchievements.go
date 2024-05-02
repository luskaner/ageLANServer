package achievement

import (
	"aoe2DELanServer/asset"
	"github.com/gin-gonic/gin"
)

func GetAvailableAchievements(c *gin.Context) {
	asset.ReturnSignedAsset("achievements.json", c, false)
}
