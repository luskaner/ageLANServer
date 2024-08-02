package advertisement

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func FindAdvertisements(w http.ResponseWriter, _ *http.Request) {
	// Only used to return online games so return nothing
	i.JSON(&w,
		i.A{0, i.A{}, i.A{}},
	)
}
