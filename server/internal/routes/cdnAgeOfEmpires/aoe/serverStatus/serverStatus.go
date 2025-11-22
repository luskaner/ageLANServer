package serverStatus

import "net/http"

func ServerStatus(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
