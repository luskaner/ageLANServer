package clan

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func Create(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{2, nil, nil, i.A{}})
}
