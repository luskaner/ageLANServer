package achievement

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/session"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAchievements(c *gin.Context) {
	sessAny, _ := c.Get("session")
	sess := sessAny.(*session.Info)
	c.JSON(http.StatusOK,
		j.A{
			0,
			j.A{
				j.A{
					sess.GetUser().GetId(),
					// DO NOT RETURN ACHIEVEMENTS AS IT WILL *REALLY* GRANT THEM ON XBOX
					j.A{},
					// asset.Achievements,
				},
			},
		},
	)
}
