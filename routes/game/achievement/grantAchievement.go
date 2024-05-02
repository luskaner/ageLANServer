package achievement

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func GrantAchievement(c *gin.Context) {
	// DO NOT ALLOW THE CLIENT TO CLAIM ACHIEVEMENTS
	c.JSON(http.StatusOK,
		i.A{
			2,
			time.Now().UTC().Unix(),
		},
	)
}
