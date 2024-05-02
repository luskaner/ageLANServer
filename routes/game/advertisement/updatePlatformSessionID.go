package advertisement

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func UpdatePlatformSessionID(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w,
		i.A{0},
	)
}
