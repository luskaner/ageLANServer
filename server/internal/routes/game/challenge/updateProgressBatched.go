package challenge

import (
	"net/http"

	i "github.com/luskaner/ageServer/server/internal"
)

func UpdateProgressBatched(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0})
}
