package item

import (
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/models"
)

func GetItemBundleItemsJson(w http.ResponseWriter, r *http.Request) {
	models.G(r).Resources().ReturnSignedAsset("itemBundleItems.json", &w, r, true)
}
