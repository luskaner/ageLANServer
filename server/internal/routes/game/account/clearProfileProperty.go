package account

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func ClearProfileProperty(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0})
}
