package relationship

import (
	"aoe2DELanServer/files"
	"net/http"
)

func GetPresenceData(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("presenceData.json", &w, r, false)
}
