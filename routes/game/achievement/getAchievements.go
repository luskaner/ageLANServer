package achievement

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAchievements(c *gin.Context) {
	sessAny, _ := c.Get("session")
	sess := sessAny.(*models.Info)
	c.JSON(http.StatusOK,
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
