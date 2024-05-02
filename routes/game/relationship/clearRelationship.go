package relationship

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func ClearRelationship(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement just in memory?
	i.JSON(&w, i.A{0})
}
