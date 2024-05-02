package achievement

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/middleware"
	"net/http"
)

func GetAchievements(w http.ResponseWriter, r *http.Request) {
	sess, _ := middleware.Session(r)
	i.JSON(&w,
		i.A{
			0,
			i.A{
				i.A{
					sess.GetUser().GetId(),
					// DO NOT RETURN ACHIEVEMENTS AS IT WILL *REALLY* GRANT THEM ON XBOX
					i.A{},
					// asset.Achievements,
				},
			},
		},
	)
}
