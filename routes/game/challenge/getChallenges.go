package challenge

import (
	"aoe2DELanServer/files"
	"net/http"
)

func GetChallenges(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("challenges.json", &w, r, false)
}
