package party

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func FinalizeReplayUpload(c *gin.Context) {
	// TODO: Implement? in memory-only does not make much sense
	c.JSON(http.StatusOK, j.A{0})
}
