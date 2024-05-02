package party

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func FinalizeReplayUpload(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement? in memory-only does not make much sense
	i.JSON(&w, i.A{0})
}
