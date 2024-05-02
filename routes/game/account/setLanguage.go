package account

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SetLanguage(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, j.A{2})
}
