package item

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func GetItemPrices(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, i.A{}, 0, i.A{}})
}
