package clan

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func Find(w http.ResponseWriter, _ *http.Request) {
	// FIXME: Try to avoid the client constinuously calling this endpoint like there were endless pages
	i.JSON(&w, i.A{0, i.A{}})
}
