package item

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SignItems(c *gin.Context) {
	// FIXME: Implement, signature seems to be base64 encoded then encrypted
	c.JSON(http.StatusOK, i.A{2, ""})
}
