package party

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ReportMatch(c *gin.Context) {
	/*advertisementIdStr := c.PostForm("match_id")
	if advertisementIdStr != "" {
		advertisementId, err := strconv.ParseUint(advertisementIdStr, 10, 32)
		if err != nil {
			adv, ok := advExtra.Get(uint32(advertisementId))
			if ok {
				adv.Delete()
			}
		}
	}*/
	// What else is needed to implement?
	c.JSON(http.StatusOK, i.A{2, i.A{}, i.A{}, i.A{}, nil, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, 0, nil, i.A{}, i.A{}, i.A{}})
}
