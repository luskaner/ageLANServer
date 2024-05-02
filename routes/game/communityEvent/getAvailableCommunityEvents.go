package communityEvent

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAvailableCommunityEvents(c *gin.Context) {
	// TODO: Implement? What is this?
	c.JSON(http.StatusOK, i.A{0, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}})
}
