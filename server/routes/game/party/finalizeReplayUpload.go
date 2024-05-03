package party

import (
	"net/http"
	i "server/internal"
)

func FinalizeReplayUpload(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement? in memory-only does not make much sense
	i.JSON(&w, i.A{0})
}
