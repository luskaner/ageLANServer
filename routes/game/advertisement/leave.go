package advertisement

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/routes/game/advertisement/extra"
	"aoe2DELanServer/session"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func Leave(c *gin.Context) {
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*session.Info)
	advStr := c.PostForm("advertisementid")
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
	currentUser := sess.GetUser()
	_, isPeer := adv.GetPeer(currentUser)
	if !isPeer {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	adv.RemovePeer(currentUser)
	c.JSON(http.StatusOK,
		j.A{0},
	)
}
