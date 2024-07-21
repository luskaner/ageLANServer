package party

import (
	"net/http"
	i "server/internal"
)

func FinalizeReplayUpload(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0})
}
