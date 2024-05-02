package clan

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Create(c *gin.Context) {
	// TODO: Implement in memory?
	c.JSON(http.StatusOK, i.A{2, nil, nil, i.A{}})
}
