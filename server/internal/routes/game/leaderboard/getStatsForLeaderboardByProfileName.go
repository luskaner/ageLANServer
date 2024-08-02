package leaderboard

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/leaderboard/shared"
	"net/http"
)

func GetStatsForLeaderboardByProfileName(w http.ResponseWriter, r *http.Request) {
	response := shared.GetStatGroups(r.URL.Query().Get("profileids"), true, false)
	i.JSON(&w, response)
}
