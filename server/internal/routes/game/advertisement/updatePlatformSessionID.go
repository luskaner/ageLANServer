package advertisement

import (
	"net/http"
)

func UpdatePlatformSessionID(w http.ResponseWriter, r *http.Request) {
	updatePlatformID(&w, r, "platformSessionID", r.FormValue("onlinePlatform"))
}
