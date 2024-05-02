package advertisement

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func Leave(c *gin.Context) {
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*models.Info)
	advStr := c.PostForm("advertisementid")
	advId, err := strconv.ParseUint(advStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	adv, ok := models.GetAdvertisement(uint32(advId))
	if !ok {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	currentUser := sess.GetUser()
	_, isPeer := adv.GetPeer(currentUser)
	if !isPeer {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	adv.RemovePeer(currentUser)
	c.JSON(http.StatusOK,
		i.A{0},
	)
}
