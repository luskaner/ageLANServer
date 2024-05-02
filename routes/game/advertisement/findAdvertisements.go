package advertisement

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func FindAdvertisements(c *gin.Context) {
	// Only used to return online games so return nothing
	c.JSON(http.StatusOK,
		i.A{0, i.A{}, i.A{}},
	)
}
