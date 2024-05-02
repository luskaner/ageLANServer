package item

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetItemLoadouts(c *gin.Context) {
	// TODO: Implement, what is this? maybe mods?
	c.JSON(http.StatusOK, j.A{0, j.A{}})
}
