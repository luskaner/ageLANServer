package advertisement

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"aoe2DELanServer/routes/game/advertisement/shared"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Update(c *gin.Context) {
	var q *shared.AdvertisementUpdateRequest
	if err := c.ShouldBind(&q); err != nil {
		c.JSON(http.StatusOK, i.A{2, i.A{}})
		return
	}
	advOriginal, ok := models.GetAdvertisement(uint32(q.Id))
	if !ok {
		c.JSON(http.StatusOK, i.A{2, i.A{}})
		return
	}
	advOriginal.Update(q)
	c.JSON(http.StatusOK,
		i.A{
			0,
			advOriginal.Encode(),
		},
	)
}
