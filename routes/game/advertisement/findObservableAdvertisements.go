package advertisement

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func FindObservableAdvertisements(w http.ResponseWriter, _ *http.Request) {
	// Return nothing as LAN games cannot be observed
	i.JSON(&w,
		i.A{0, i.A{}, i.A{}},
	)
}
