package achievement

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func SyncStats(w http.ResponseWriter, _ *http.Request) {
	// What does it do?
	i.JSON(&w,
		i.A{2},
	)
}
