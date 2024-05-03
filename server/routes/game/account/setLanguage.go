package account

import (
	"net/http"
	i "server/internal"
)

func SetLanguage(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement
	i.JSON(&w, i.A{2})
}
