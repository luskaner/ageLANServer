package leaderboard

import (
	"net/http"
	i "server/internal"
	"server/routes/game/leaderboard/shared"
)

func GetStatsForLeaderboardByProfileName(w http.ResponseWriter, r *http.Request) {
	response := shared.GetStatGroups(r.URL.Query().Get("profileids"), true, false)
	i.JSON(&w, response)
}
