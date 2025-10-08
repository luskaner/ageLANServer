package account

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func GetProfileProperty(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, i.A{}})
}
