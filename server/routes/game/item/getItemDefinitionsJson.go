package item

import (
	"net/http"
	"server/files"
)

func GetItemDefinitionsJson(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("itemDefinitions.json", &w, r, true)
}
