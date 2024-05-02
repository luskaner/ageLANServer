package advertisement

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAdvertisements(c *gin.Context) {
	matchIdsStr := c.Query("match_ids")
	var advsIds []uint32
	err := json.Unmarshal([]byte(matchIdsStr), &advsIds)
	if err != nil {
		c.JSON(http.StatusOK, i.A{2, i.A{}})
		return
	}
	advs := models.FindAdvertisementsEncoded(func(adv *models.Advertisement) bool {
		for _, advId := range advsIds {
			if adv.GetId() == advId {
				return true
			}
		}
		return false
	})
	if advs == nil {
		c.JSON(http.StatusOK,
			i.A{0, i.A{}},
		)
	} else {
		c.JSON(http.StatusOK,
			i.A{0, advs},
		)
	}
}
