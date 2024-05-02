package challenge

import (
	"aoe2DELanServer/files"
	"github.com/gin-gonic/gin"
)

func GetChallenges(c *gin.Context) {
	files.ReturnSignedAsset("challenges.json", c, false)
}
