package communityEvent

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func GetEventLeaderboard(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{2, nil, i.A{}, i.A{}, i.A{}})
}
