package achievement

import (
	"net/http"
	i "server/internal"
)

func SyncStats(w http.ResponseWriter, _ *http.Request) {
	// TODO: What does it do?
	i.JSON(&w,
		i.A{2},
	)
}
