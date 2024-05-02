package advertisement

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func FindAdvertisements(c *gin.Context) {
	// Only used to return online games so return nothing
	c.JSON(http.StatusOK,
		j.A{0, j.A{}, j.A{}},
	)
}
