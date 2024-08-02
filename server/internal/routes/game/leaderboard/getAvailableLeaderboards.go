package leaderboard

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/files"
	"net/http"
)

func GetAvailableLeaderboards(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, files.ArrayFiles["leaderboards.json"])
}
