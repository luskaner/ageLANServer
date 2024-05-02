package relationship

import (
	"aoe2DELanServer/asset"
	"github.com/gin-gonic/gin"
)

func GetPresenceData(c *gin.Context) {
	asset.ReturnSignedAsset("presence_data.json", c, false)
}
