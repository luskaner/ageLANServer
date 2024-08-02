package item

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/files"
	"net/http"
)

func GetItemDefinitionsJson(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("itemDefinitions.json", &w, r, true)
}
