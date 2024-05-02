package item

import (
	"aoe2DELanServer/asset"
	"github.com/gin-gonic/gin"
)

func GetItemDefinitionsJson(c *gin.Context) {
	asset.ReturnSignedAsset("item_definitions.json", c, true)
}
