package clan

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Find(c *gin.Context) {
	// FIXME: Try to avoid the client constinuously calling this endpoint like there were endless pages
	c.JSON(http.StatusOK, j.A{0, j.A{}})
}
