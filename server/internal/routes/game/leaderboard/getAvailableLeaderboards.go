package leaderboard

import (
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

func GetAvailableLeaderboards(w http.ResponseWriter, r *http.Request) {
	models.G(r).Resources().ReturnArrayFile("leaderboards.json", &w)
}
