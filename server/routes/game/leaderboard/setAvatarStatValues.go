package leaderboard

import (
	"net/http"
	i "server/internal"
)

func SetAvatarStatValues(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement?
	i.JSON(&w, i.A{0})
}
