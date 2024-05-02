package advertisement

import (
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func UpdateTags(c *gin.Context) {
	/*advStr := c.PostForm("advertisementid")
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
	tagNamesStr := c.PostForm("numericTagNames")
	var tagNames []string
	err = json.Unmarshal([]byte(tagNamesStr), &tagNames)
	if err != nil {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	tagValuesStr := c.PostForm("numericTagValues")
	var tagValues []int32
	err = json.Unmarshal([]byte(tagValuesStr), &tagValues)
	if err != nil {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	tags := make(map[string]int32)
	for i, key := range tagNames {
		tags[key] = tagValues[i]
	}
	adv.UpdateTags(tags)*/
	// Only used by findAdvertisements which does not return LAN games
	c.JSON(http.StatusOK, i.A{0})
}
