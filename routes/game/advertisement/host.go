package advertisement

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/routes/game/advertisement/extra"
	"aoe2DELanServer/user"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

var re *regexp.Regexp = nil

func returnError(c *gin.Context) {
	c.JSON(http.StatusOK, j.A{
		2,
		0,
		"",
		"",
		0,
		0,
		0,
		"",
		j.A{},
		0,
		0,
		nil,
		nil,
		"",
		"",
	})
}

func Host(c *gin.Context) {
	if re == nil {
		re, _ = regexp.Compile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[89aAbB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)
	}

	if !re.MatchString(c.PostForm("relayRegion")) {
		returnError(c)
		return
	}

	var adv extra.AdvertisementHostRequest
	if err := c.ShouldBind(&adv); err == nil {
		u, ok := user.GetById(adv.HostId)
		if !ok || extra.IsInAdvertisement(u) {
			returnError(c)
			return
		}
		storedAdv := extra.Store(&adv)
		c.JSON(http.StatusOK,
			j.A{
				0,
				storedAdv.GetId(),
				"authtoken",
				"",
				0,
				0,
				0,
				storedAdv.GetRelayRegion(),
				storedAdv.EncodePeers(),
				0,
				0,
				nil,
				nil,
				"0",
				storedAdv.GetDescription(),
			},
		)
	} else {
		returnError(c)
	}
}
