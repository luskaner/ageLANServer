package leaderboard

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func GetLeaderBoard(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, i.A{}, i.A{}, i.A{}})
}
