package challenge

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"net/http"
)

func GetChallenges(w http.ResponseWriter, r *http.Request) {
	models.G(r).Resources().ReturnSignedAsset("challenges.json", &w, r, false)
}
