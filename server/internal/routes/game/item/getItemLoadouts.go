package item

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func GetItemLoadouts(w http.ResponseWriter, _ *http.Request) {
	// What is this? maybe mods?
	i.JSON(&w, i.A{0, i.A{}})
}
