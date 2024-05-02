package account

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SetAvatarMetadata(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, i.A{2, i.A{}})
}
