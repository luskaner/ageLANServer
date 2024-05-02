package relationship

import (
	"aoe2DELanServer/files"
	"github.com/gin-gonic/gin"
)

func GetPresenceData(c *gin.Context) {
	files.ReturnSignedAsset("presenceData.json", c, false)
}
