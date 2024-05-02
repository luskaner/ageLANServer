package party

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func UpdateHost(c *gin.Context) {
	advStr := c.PostForm("match_id")
	advId, err := strconv.ParseUint(advStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	adv, ok := models.GetAdvertisement(uint32(advId))
	if !ok {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	ok = adv.UpdateHost()
	var code int
	if ok {
		code = 1
	} else {
		code = 2
	}
	c.JSON(http.StatusOK, i.A{code})
}
