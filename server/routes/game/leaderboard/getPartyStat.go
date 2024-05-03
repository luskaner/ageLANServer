package leaderboard

import (
	"net/http"
	i "server/internal"
	"server/routes/game/leaderboard/shared"
)

func GetPartyStat(w http.ResponseWriter, r *http.Request) {
	response := shared.GetStatGroups(r.URL.Query().Get("statsids"), false, true)
	i.JSON(&w, response)
}
