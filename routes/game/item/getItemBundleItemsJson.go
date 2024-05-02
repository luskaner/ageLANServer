package item

import (
	"aoe2DELanServer/files"
	"net/http"
)

func GetItemBundleItemsJson(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("itemBundleItems.json", &w, r, true)
}
