package leaderboard

import (
	"aoe2DELanServer/files"
	i "aoe2DELanServer/internal"
	"net/http"
)

func GetAvailableLeaderboards(w http.ResponseWriter, _ *http.Request) {
	file := files.Config["leaderboards.json"]
	response := make(i.A, len(file))
	copy(response, file)
	response = append(i.A{0}, response...)
	i.JSON(&w, response)
}
