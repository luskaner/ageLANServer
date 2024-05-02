package item

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func SignItems(w http.ResponseWriter, _ *http.Request) {
	// FIXME: Implement, signature seems to be base64 encoded then encrypted
	i.JSON(&w, i.A{2, ""})
}
