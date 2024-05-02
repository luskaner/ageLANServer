package clan

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Create(c *gin.Context) {
	// TODO: Implement in memory?
	c.JSON(http.StatusOK, j.A{2, nil, nil, j.A{}})
}
