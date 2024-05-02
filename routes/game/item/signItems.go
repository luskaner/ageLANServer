package item

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SignItems(c *gin.Context) {
	// FIXME: Implement, signature seems to be base64 encoded then encrypted
	c.JSON(http.StatusOK, j.A{2, ""})
}
