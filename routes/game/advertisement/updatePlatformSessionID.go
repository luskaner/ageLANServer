package advertisement

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func UpdatePlatformSessionID(c *gin.Context) {
	c.JSON(http.StatusOK,
		j.A{0},
	)
}
