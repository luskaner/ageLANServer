package leaderboard

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/leaderboard/shared"
	"net/http"
)

func GetStatsForLeaderboardByProfileName(w http.ResponseWriter, r *http.Request) {
	response := shared.GetStatGroups(r, r.URL.Query().Get("profileids"), true, false)
	i.JSON(&w, response)
}
