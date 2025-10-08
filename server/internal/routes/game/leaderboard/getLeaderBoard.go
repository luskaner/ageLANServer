package leaderboard

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func GetLeaderBoard(w http.ResponseWriter, _ *http.Request) {
	// Check it works fine for Aoe3
	i.JSON(&w, i.A{0, i.A{}, i.A{}, i.A{}})
}
