package achievement

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"net/http"
)

func SyncStats(w http.ResponseWriter, _ *http.Request) {
	// What does it do?
	i.JSON(&w,
		i.A{2},
	)
}
