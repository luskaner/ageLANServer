package item

import (
	"github.com/luskaner/aoe2DELanServer/server/files"
	"net/http"
)

func GetItemBundleItemsJson(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("itemBundleItems.json", &w, r, true)
}
