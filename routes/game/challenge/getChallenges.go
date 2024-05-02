package challenge

import (
	"aoe2DELanServer/asset"
	"github.com/gin-gonic/gin"
)

func GetChallenges(c *gin.Context) {
	asset.ReturnSignedAsset("challenges.json", c, false)
}
