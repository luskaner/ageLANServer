package leaderboard

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func GetRecentMatchHistory(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, i.A{}})
}
