package achievement

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
)

func GetAchievements(w http.ResponseWriter, r *http.Request) {
	sess := middleware.Session(r)
	i.JSON(&w,
		i.A{
			0,
			i.A{
				i.A{
					sess.GetUserId(),
					// DO NOT RETURN ACHIEVEMENTS AS IT WILL *REALLY* GRANT THEM ON XBOX
					i.A{},
				},
			},
		},
	)
}
