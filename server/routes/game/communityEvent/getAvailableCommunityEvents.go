package communityEvent

import (
	"net/http"
	i "server/internal"
)

func GetAvailableCommunityEvents(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement? What is this?
	i.JSON(&w, i.A{0, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}})
}
