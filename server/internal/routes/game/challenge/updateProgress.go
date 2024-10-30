package challenge

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func UpdateProgress(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{2, i.A{}, i.A{}})
}
