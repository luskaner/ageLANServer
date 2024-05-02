package account

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func SetLanguage(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement
	i.JSON(&w, i.A{2})
}
