package advertisement

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func FindObservableAdvertisements(c *gin.Context) {
	// Return nothing as LAN games cannot be observed
	c.JSON(http.StatusOK,
		j.A{0, j.A{}, j.A{}},
	)
}
