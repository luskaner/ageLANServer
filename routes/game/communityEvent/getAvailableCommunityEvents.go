package communityEvent

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func GetAvailableCommunityEvents(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement? What is this?
	i.JSON(&w, i.A{0, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}})
}
