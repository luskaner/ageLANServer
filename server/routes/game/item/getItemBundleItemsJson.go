package item

import (
	"net/http"
	"server/files"
)

func GetItemBundleItemsJson(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("itemBundleItems.json", &w, r, true)
}
