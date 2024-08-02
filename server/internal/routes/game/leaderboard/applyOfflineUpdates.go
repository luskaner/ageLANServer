package leaderboard

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func ApplyOfflineUpdates(w http.ResponseWriter, _ *http.Request) {
	// Which kind of updates?
	i.JSON(&w, i.A{0, i.A{}, i.A{}})
}
