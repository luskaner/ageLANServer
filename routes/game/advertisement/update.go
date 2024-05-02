package advertisement

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/routes/game/advertisement/extra"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Update(c *gin.Context) {
	var q *extra.AdvertisementUpdateRequest
	if err := c.ShouldBind(&q); err != nil {
		c.JSON(http.StatusOK, j.A{2, j.A{}})
		return
	}
	advOriginal, ok := extra.Get(uint32(q.Id))
	if !ok {
		c.JSON(http.StatusOK, j.A{2, j.A{}})
		return
	}
	advOriginal.Update(q)
	c.JSON(http.StatusOK,
		j.A{
			0,
			advOriginal.Encode(),
		},
	)
}
