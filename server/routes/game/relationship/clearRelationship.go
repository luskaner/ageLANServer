package relationship

import (
	"net/http"
	i "server/internal"
)

func ClearRelationship(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0})
}
