package relationship

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ClearRelationship(c *gin.Context) {
	// TODO: Implement just in memory?
	c.JSON(http.StatusOK, i.A{0})
}
