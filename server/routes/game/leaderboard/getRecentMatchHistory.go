package leaderboard

import (
	"net/http"
	i "server/internal"
)

func GetRecentMatchHistory(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, i.A{}})
}
