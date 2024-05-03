package challenge

import (
	"net/http"
	"server/files"
)

func GetChallenges(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("challenges.json", &w, r, false)
}
