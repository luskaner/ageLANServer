package leaderboard

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"net/http"
)

func GetLeaderBoard(w http.ResponseWriter, _ *http.Request) {
	// Check it works fine for Aoe3
	i.JSON(&w, i.A{0, i.A{}, i.A{}, i.A{}})
}
