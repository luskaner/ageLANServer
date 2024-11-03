package party

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"net/http"
)

func FinalizeReplayUpload(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0})
}
