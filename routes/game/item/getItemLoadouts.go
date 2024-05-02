package item

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetItemLoadouts(c *gin.Context) {
	// TODO: Implement, what is this? maybe mods?
	c.JSON(http.StatusOK, i.A{0, i.A{}})
}
