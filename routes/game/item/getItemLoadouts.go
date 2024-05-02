package item

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func GetItemLoadouts(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement, what is this? maybe mods?
	i.JSON(&w, i.A{0, i.A{}})
}
