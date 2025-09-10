package advertisement

import (
	"net/http"
)

func UpdatePlatformSessionID(w http.ResponseWriter, r *http.Request) {
	// TODO: Use "onlinePlatform" - STEAM... ?
	updatePlatformID(&w, r, "platformSessionID")
}
