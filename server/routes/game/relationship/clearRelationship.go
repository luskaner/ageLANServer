package relationship

import (
	"net/http"
	i "server/internal"
)

func ClearRelationship(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement just in memory?
	i.JSON(&w, i.A{0})
}
