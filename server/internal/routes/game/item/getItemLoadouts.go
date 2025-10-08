package item

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func GetItemLoadouts(w http.ResponseWriter, _ *http.Request) {
	// What is this? maybe mods?
	i.JSON(&w, i.A{0, i.A{}})
}
