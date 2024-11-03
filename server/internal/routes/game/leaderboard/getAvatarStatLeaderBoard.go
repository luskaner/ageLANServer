package leaderboard

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"net/http"
)

func GetAvatarStatLeaderBoard(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, i.A{}, i.A{}})
}
