package item

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func GetPersonalizedSaleItems(w http.ResponseWriter, r *http.Request) {
	i.JSON(&w, i.A{0, i.A{}, i.A{}, 0})
}
