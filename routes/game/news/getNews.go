package news

import (
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetNews(c *gin.Context) {
	c.JSON(http.StatusOK, j.A{0, j.A{}, j.A{}})
}
