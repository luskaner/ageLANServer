package advertisement

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"net/http"
)

func UpdateTags(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0})
}
