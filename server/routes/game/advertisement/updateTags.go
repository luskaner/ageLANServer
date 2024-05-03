package advertisement

import (
	"net/http"
	i "server/internal"
)

func UpdateTags(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0})
}
