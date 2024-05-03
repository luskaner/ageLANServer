package achievement

import (
	"net/http"
	i "server/internal"
	"time"
)

func GrantAchievement(w http.ResponseWriter, _ *http.Request) {
	// DO NOT ALLOW THE CLIENT TO CLAIM ACHIEVEMENTS
	i.JSON(&w,
		i.A{
			2,
			time.Now().UTC().Unix(),
		},
	)
}
