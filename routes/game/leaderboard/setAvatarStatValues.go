package leaderboard

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func SetAvatarStatValues(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement?
	i.JSON(&w, i.A{0})
}
