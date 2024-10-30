package item

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"net/http"
)

func GetItemDefinitionsJson(w http.ResponseWriter, r *http.Request) {
	models.G(r).Resources().ReturnSignedAsset("itemDefinitions.json", &w, r, true)
}
