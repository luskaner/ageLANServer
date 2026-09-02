package Automatch2

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func Stoppolling(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement for matchmaking
	i.JSON(&w, i.A{0})
}
