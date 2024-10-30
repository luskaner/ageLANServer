package advertisement

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func UpdatePlatformSessionID(w http.ResponseWriter, _ *http.Request) {
	// Unknown what's used for
	i.JSON(&w,
		i.A{0},
	)
}
