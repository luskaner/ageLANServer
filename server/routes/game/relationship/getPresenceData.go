package relationship

import (
	"net/http"
	"server/files"
)

func GetPresenceData(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("presenceData.json", &w, r, false)
}
