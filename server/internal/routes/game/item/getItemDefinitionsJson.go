package item

import (
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/models"
)

func GetItemDefinitionsJson(w http.ResponseWriter, r *http.Request) {
	models.G(r).Resources().ReturnSignedAsset("itemDefinitions.json", &w, r, true)
}
