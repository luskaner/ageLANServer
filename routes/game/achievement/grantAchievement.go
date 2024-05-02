package achievement

import (
	i "aoe2DELanServer/internal"
	"net/http"
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
