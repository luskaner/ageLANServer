package item

import (
	"aoe2DELanServer/files"
	"github.com/gin-gonic/gin"
)

func GetItemDefinitionsJson(c *gin.Context) {
	files.ReturnSignedAsset("itemDefinitions.json", c, true)
}
