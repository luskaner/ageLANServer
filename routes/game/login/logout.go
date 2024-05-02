package login

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Logout(c *gin.Context) {
	sessAny, _ := c.Get("session")
	sess := sessAny.(*models.Info)
	u := sess.GetUser()
	advs := models.FindAdvertisements(func(adv *models.Advertisement) bool {
		_, found := adv.GetPeer(u)
		return found
	})
	for _, adv := range advs {
		adv.RemovePeer(u)
	}
	u.SetPresence(0)
	sess.Delete()
	c.JSON(http.StatusOK, i.A{0})
}
