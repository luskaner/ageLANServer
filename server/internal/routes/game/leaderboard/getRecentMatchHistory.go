package leaderboard

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/logger"
)

func GetRecentMatchHistory(w http.ResponseWriter, r *http.Request) {
	logger.Printf("leaderboard.getRecentMatchHistory: query=%s\n", r.URL.RawQuery)
	i.JSON(&w, i.A{0, i.A{}})
}
