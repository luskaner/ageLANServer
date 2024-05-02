package msstore

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetStoreTokens(c *gin.Context) {
	// Likely just used to then send through platformlogin, is it for DLCs?
	c.JSON(http.StatusOK, i.A{0, nil, ""})
}
