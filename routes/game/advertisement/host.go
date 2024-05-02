package advertisement

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"aoe2DELanServer/routes/game/advertisement/shared"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

var re *regexp.Regexp = nil

func returnError(c *gin.Context) {
	c.JSON(http.StatusOK, i.A{
		2,
		0,
		"",
		"",
		0,
		0,
		0,
		"",
		i.A{},
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

	var adv shared.AdvertisementHostRequest
	if err := c.ShouldBind(&adv); err == nil {
		u, ok := models.GetUserById(adv.HostId)
		if !ok || models.IsInAdvertisement(u) {
			returnError(c)
			return
		}
		storedAdv := models.StoreAdvertisement(&adv)
		c.JSON(http.StatusOK,
			i.A{
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
