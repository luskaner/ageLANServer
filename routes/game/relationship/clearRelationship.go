package relationship

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ClearRelationship(c *gin.Context) {
	// TODO: Implement just in memory?
	c.JSON(http.StatusOK, j.A{0})
}
