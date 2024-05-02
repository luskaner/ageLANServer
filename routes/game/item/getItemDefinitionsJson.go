package item

import (
	"aoe2DELanServer/files"
	"net/http"
)

func GetItemDefinitionsJson(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("itemDefinitions.json", &w, r, true)
}
