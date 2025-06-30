package item

import (
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

func GetItemBundleItemsJson(w http.ResponseWriter, r *http.Request) {
	models.G(r).Resources().ReturnSignedAsset("itemBundleItems.json", &w, r)
}
