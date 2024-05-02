package msstore

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetStoreTokens(c *gin.Context) {
	// Likely just used to then send through platformlogin, is it for DLCs?
	c.JSON(http.StatusOK, j.A{0, nil, ""})
}
