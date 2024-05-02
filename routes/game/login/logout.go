package login

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/routes/game/advertisement/extra"
	"aoe2DELanServer/session"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Logout(c *gin.Context) {
	sessAny, _ := c.Get("session")
	sess := sessAny.(*session.Info)
	u := sess.GetUser()
	advs := extra.FindAdvertisementsOriginal(func(adv *extra.Advertisement) bool {
		_, found := adv.GetPeer(u)
		return found
	})
	for _, adv := range advs {
		adv.RemovePeer(u)
	}
	u.SetPresence(0)
	sess.Delete()
	c.JSON(http.StatusOK, j.A{0})
}
