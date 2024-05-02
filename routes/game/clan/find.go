package clan

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Find(c *gin.Context) {
	// FIXME: Try to avoid the client constinuously calling this endpoint like there were endless pages
	c.JSON(http.StatusOK, i.A{0, i.A{}})
}
