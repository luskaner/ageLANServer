package party

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func FinalizeReplayUpload(c *gin.Context) {
	// TODO: Implement? in memory-only does not make much sense
	c.JSON(http.StatusOK, i.A{0})
}
