package communityEvent

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAvailableCommunityEvents(c *gin.Context) {
	// TODO: Implement? What is this?
	c.JSON(http.StatusOK, j.A{0, j.A{}, j.A{}, j.A{}, j.A{}, j.A{}, j.A{}})
}
