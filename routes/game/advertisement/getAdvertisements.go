package advertisement

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/routes/game/advertisement/extra"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAdvertisements(c *gin.Context) {
	matchIdsStr := c.Query("match_ids")
	var advsIds []uint32
	err := json.Unmarshal([]byte(matchIdsStr), &advsIds)
	if err != nil {
		c.JSON(http.StatusOK, j.A{2, j.A{}})
		return
	}
	advs := extra.FindAdvertisementsEncoded(func(adv *extra.Advertisement) bool {
		for _, advId := range advsIds {
			if adv.GetId() == advId {
				return true
			}
		}
		return false
	})
	if advs == nil {
		c.JSON(http.StatusOK,
			j.A{0, j.A{}},
		)
	} else {
		c.JSON(http.StatusOK,
			j.A{0, advs},
		)
	}
}
