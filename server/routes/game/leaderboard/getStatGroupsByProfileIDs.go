package leaderboard

import (
	"net/http"
	i "server/internal"
	"server/routes/game/leaderboard/shared"
)

func GetStatGroupsByProfileIDs(w http.ResponseWriter, r *http.Request) {
	response := shared.GetStatGroups(r.URL.Query().Get("profileids"), true, true)
	i.JSON(&w, response)
}
