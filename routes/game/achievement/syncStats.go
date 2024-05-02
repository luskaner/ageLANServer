package achievement

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func SyncStats(w http.ResponseWriter, _ *http.Request) {
	// TODO: What does it do?
	i.JSON(&w,
		i.A{2},
	)
}
