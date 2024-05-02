package relationship

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func Ignore(c *gin.Context) {
	// TODO: Implement just in memory?
	profileIdStr := c.PostForm("targetProfileID")
	profileId, err := strconv.Atoi(profileIdStr)
	if err != nil {
		c.JSON(http.StatusOK, i.A{2, i.A{}, i.A{}})
		return
	}
	u, ok := models.GetUserById(int32(profileId))
	if !ok {
		c.JSON(http.StatusOK, i.A{2, i.A{}, i.A{}})
		return
	}
	c.JSON(http.StatusOK, i.A{2, u.GetProfileInfo(false), i.A{}})
}
