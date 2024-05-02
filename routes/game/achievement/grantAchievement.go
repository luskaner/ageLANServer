package achievement

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func GrantAchievement(c *gin.Context) {
	// DO NOT ALLOW THE CLIENT TO CLAIM ACHIEVEMENTS
	c.JSON(http.StatusOK,
		j.A{
			2,
			time.Now().UTC().Unix(),
		},
	)
}
