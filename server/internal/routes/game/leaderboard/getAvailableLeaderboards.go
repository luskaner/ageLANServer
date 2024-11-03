package leaderboard

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

func GetAvailableLeaderboards(w http.ResponseWriter, r *http.Request) {
	i.JSON(&w, models.G(r).Resources().ArrayFiles["leaderboards.json"])
}
