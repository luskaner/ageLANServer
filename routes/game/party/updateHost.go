package party

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/routes/game/advertisement/extra"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func UpdateHost(c *gin.Context) {
	advStr := c.PostForm("match_id")
	advId, err := strconv.ParseUint(advStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	adv, ok := extra.Get(uint32(advId))
	if !ok {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	ok = adv.UpdateHost()
	var code int
	if ok {
		code = 1
	} else {
		code = 2
	}
	c.JSON(http.StatusOK, j.A{code})
}
