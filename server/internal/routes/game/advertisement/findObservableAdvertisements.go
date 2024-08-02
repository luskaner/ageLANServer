package advertisement

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func FindObservableAdvertisements(w http.ResponseWriter, _ *http.Request) {
	// Return nothing as LAN games cannot be observed
	i.JSON(&w,
		i.A{0, i.A{}, i.A{}},
	)
}
