package leaderboard

import (
	"net/http"
	"server/files"
	i "server/internal"
)

func GetAvailableLeaderboards(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, files.ArrayFiles["leaderboards.json"])
}
