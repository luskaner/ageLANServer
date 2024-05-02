package advertisement

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func FindObservableAdvertisements(c *gin.Context) {
	// Return nothing as LAN games cannot be observed
	c.JSON(http.StatusOK,
		i.A{0, i.A{}, i.A{}},
	)
}
