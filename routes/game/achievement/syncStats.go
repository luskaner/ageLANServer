package achievement

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SyncStats(c *gin.Context) {
	// TODO: What does it do?
	c.JSON(http.StatusOK,
		i.A{2},
	)
}
