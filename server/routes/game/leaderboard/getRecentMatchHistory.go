package leaderboard

import (
	"net/http"
	i "server/internal"
)

func GetRecentMatchHistory(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement? just in memory does not make much sense
	i.JSON(&w, i.A{0, i.A{}})
}
