package party

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"aoe2DELanServer/routes/game/party/shared"
	"github.com/gin-gonic/gin"
	"net/http"
)

func PeerAdd(c *gin.Context) {
	adv, length, profileIds, raceIds, statGroupIds, teamIds := shared.ParseParameters(c)
	if adv == nil {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*models.Info)
	currentUser := sess.GetUser()
	// Only the host can add peers
	host := adv.GetHost().GetUser()
	if host != currentUser {
		c.JSON(http.StatusOK, i.A{2})
		return
	}
	users := make([]*models.User, length)
	for j := 0; j < length; j++ {
		u, ok := models.GetUserById(profileIds[j])
		if !ok || u.GetStatId() != statGroupIds[j] {
			c.JSON(http.StatusOK, i.A{2})
			return
		}
		users[j] = u
	}
	for j, u := range users {
		adv.NewPeer(u, raceIds[j], teamIds[j])
	}
	c.JSON(http.StatusOK, i.A{0})
}
