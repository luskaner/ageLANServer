package party

import (
	"aoe2DELanServer/j"
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
	c.JSON(http.StatusOK, j.A{2, j.A{}, j.A{}, j.A{}, nil, j.A{}, j.A{}, j.A{}, j.A{}, j.A{}, j.A{}, 0, nil, j.A{}, j.A{}, j.A{}})
}
