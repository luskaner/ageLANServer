package account

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"net/http"
)

func SetLanguage(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{2})
}
