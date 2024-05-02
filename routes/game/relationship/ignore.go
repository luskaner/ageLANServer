package relationship

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/user"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func Ignore(c *gin.Context) {
	// TODO: Implement just in memory?
	profileIdStr := c.PostForm("targetProfileID")
	profileId, err := strconv.Atoi(profileIdStr)
	if err != nil {
		c.JSON(http.StatusOK, j.A{2, j.A{}, j.A{}})
		return
	}
	u, ok := user.GetById(int32(profileId))
	if !ok {
		c.JSON(http.StatusOK, j.A{2, j.A{}, j.A{}})
		return
	}
	c.JSON(http.StatusOK, j.A{2, u.GetProfileInfo(false), j.A{}})
}
