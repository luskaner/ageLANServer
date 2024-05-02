package party

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/routes/game/party/extra"
	"aoe2DELanServer/session"
	"aoe2DELanServer/user"
	"github.com/gin-gonic/gin"
	"net/http"
)

func PeerAdd(c *gin.Context) {
	adv, length, profileIds, raceIds, statGroupIds, teamIds := extra.ParseParameters(c)
	if adv == nil {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*session.Info)
	currentUser := sess.GetUser()
	// Only the host can add peers
	host := adv.GetHost().GetUser()
	if host != currentUser {
		c.JSON(http.StatusOK, j.A{2})
		return
	}
	users := make([]*user.User, length)
	for i := 0; i < length; i++ {
		u, ok := user.GetById(profileIds[i])
		if !ok || u.GetStatId() != statGroupIds[i] {
			c.JSON(http.StatusOK, j.A{2})
			return
		}
		users[i] = u
	}
	for i, u := range users {
		adv.NewPeer(u, raceIds[i], teamIds[i])
	}
	c.JSON(http.StatusOK, j.A{0})
}
