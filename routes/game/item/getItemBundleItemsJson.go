package item

import (
	"aoe2DELanServer/asset"
	"github.com/gin-gonic/gin"
)

func GetItemBundleItemsJson(c *gin.Context) {
	asset.ReturnSignedAsset("item_bundle_items.json", c, true)
}
