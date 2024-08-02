package leaderboard

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func GetRecentMatchHistory(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, i.A{}})
}
