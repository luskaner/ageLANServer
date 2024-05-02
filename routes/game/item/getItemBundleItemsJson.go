package item

import (
	"aoe2DELanServer/files"
	"github.com/gin-gonic/gin"
)

func GetItemBundleItemsJson(c *gin.Context) {
	files.ReturnSignedAsset("itemBundleItems.json", c, true)
}
