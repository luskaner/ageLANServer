package account

import (
	"net/http"
	i "server/internal"
)

func SetCrossplayEnabled(w http.ResponseWriter, r *http.Request) {
	// Crossplay is always enabled regardless of the value sent
	enable := r.PostFormValue("enable")
	if enable == "1" {
		i.JSON(&w, i.A{0})
	} else {
		// Do not accept disabling it
		i.JSON(&w, i.A{2})
	}
}
