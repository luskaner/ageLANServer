package communityEvent

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func GetAvailableCommunityEvents(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement? What is this?
	i.JSON(&w, i.A{0, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}})
}
