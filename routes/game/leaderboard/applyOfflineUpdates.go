package leaderboard

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func ApplyOfflineUpdates(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement? which kind of updates?
	i.JSON(&w, i.A{0, i.A{}, i.A{}})
}
