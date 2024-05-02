package clan

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func Create(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement in memory?
	i.JSON(&w, i.A{2, nil, nil, i.A{}})
}
