package achievement

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SyncStats(c *gin.Context) {
	// TODO: What does it do?
	c.JSON(http.StatusOK,
		j.A{2},
	)
}
