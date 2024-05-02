package leaderboard

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/routes/game/leaderboard/shared"
	"net/http"
)

func GetStatGroupsByProfileIDs(w http.ResponseWriter, r *http.Request) {
	response := shared.GetStatGroups(r.URL.Query().Get("profileids"), true, true)
	i.JSON(&w, response)
}
