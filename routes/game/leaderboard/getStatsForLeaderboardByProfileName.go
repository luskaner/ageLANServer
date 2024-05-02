package leaderboard

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/routes/game/leaderboard/shared"
	"net/http"
)

func GetStatsForLeaderboardByProfileName(w http.ResponseWriter, r *http.Request) {
	response := shared.GetStatGroups(r.URL.Query().Get("profileids"), true, false)
	i.JSON(&w, response)
}
