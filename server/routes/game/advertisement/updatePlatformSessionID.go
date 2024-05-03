package advertisement

import (
	"net/http"
	i "server/internal"
)

func UpdatePlatformSessionID(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w,
		i.A{0},
	)
}
