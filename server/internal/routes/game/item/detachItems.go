package item

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"net/http"
)

func DetachItems(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{2, i.A{0}, i.A{}})
}
