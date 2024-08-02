package challenge

import (
	"github.com/luskaner/aoe2DELanServer/server/files"
	"net/http"
)

func GetChallenges(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("challenges.json", &w, r, false)
}
