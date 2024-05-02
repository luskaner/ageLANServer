package msstore

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func GetStoreTokens(w http.ResponseWriter, r *http.Request) {
	// Likely just used to then send through platformlogin, is it for DLCs?
	i.JSON(&w, i.A{0, nil, ""})
}
