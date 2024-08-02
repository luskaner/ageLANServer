package leaderboard

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func SetAvatarStatValues(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0})
}
