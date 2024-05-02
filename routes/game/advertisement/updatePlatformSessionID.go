package advertisement

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func UpdatePlatformSessionID(c *gin.Context) {
	c.JSON(http.StatusOK,
		i.A{0},
	)
}
