package leaderboard

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/logger"
)

func GetLeaderBoard(w http.ResponseWriter, r *http.Request) {
	logger.Printf("leaderboard.getLeaderBoard: query=%s\n", r.URL.RawQuery)
	// Check it works fine for Aoe3
	i.JSON(&w, i.A{0, i.A{}, i.A{}, i.A{}})
}
