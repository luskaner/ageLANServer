package item

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func DetachItems(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{2, i.A{0}, i.A{}})
}
