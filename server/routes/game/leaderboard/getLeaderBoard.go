package leaderboard

import (
	"net/http"
	i "server/internal"
)

func GetLeaderBoard(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, i.A{}, i.A{}, i.A{}})
}
