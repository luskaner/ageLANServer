package leaderboard

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func GetAvatarStatLeaderBoard(w http.ResponseWriter, _ *http.Request) {
	// TODO Check it works (aoe3)
	i.JSON(&w, i.A{0, i.A{}, i.A{}})
}
