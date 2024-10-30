package item

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"net/http"
)

func GetItemBundleItemsJson(w http.ResponseWriter, r *http.Request) {
	models.G(r).Resources().ReturnSignedAsset("itemBundleItems.json", &w, r, true)
}
