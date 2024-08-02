package msstore

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func GetStoreTokens(w http.ResponseWriter, _ *http.Request) {
	// Likely just used to then send through platformlogin, is it for DLCs?
	i.JSON(&w, i.A{0, nil, ""})
}
