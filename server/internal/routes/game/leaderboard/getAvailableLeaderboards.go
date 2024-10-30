package leaderboard

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"net/http"
)

func GetAvailableLeaderboards(w http.ResponseWriter, r *http.Request) {
	i.JSON(&w, models.G(r).Resources().ArrayFiles["leaderboards.json"])
}
