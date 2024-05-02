package advertisement

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"aoe2DELanServer/routes/game/advertisement/shared"
	"github.com/gin-gonic/gin"
	"net/http"
)

type JoinRequest struct {
	shared.AdvertisementBaseRequest
	Password shared.PasswordBaseRequest
}

func Join(c *gin.Context) {
	var q JoinRequest
	if err := c.ShouldBind(&q); err != nil {
		c.JSON(http.StatusOK, i.A{2, "", "", 0, 0, 0, i.A{}})
		return
	}
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*models.Info)
	u := sess.GetUser()
	if models.IsInAdvertisement(u) {
		c.JSON(http.StatusOK, i.A{2, "", "", 0, 0, 0, i.A{}})
		return
	}
	advId := uint32(q.Id)
	advs := models.FindAdvertisements(func(adv *models.Advertisement) bool {
		return adv.GetId() == advId && adv.GetJoinable() && adv.GetAppBinaryChecksum() == q.AppBinaryChecksum && adv.GetDataChecksum() == q.DataChecksum && adv.GetModDllFile() == q.ModDll.File && adv.GetModDllChecksum() == q.ModDll.Checksum && adv.GetModName() == q.ModName && adv.GetModVersion() == q.ModVersion && adv.GetVersionFlags() == q.VersionFlags && adv.GetPasswordValue() == q.Password.Value
	})
	if len(advs) != 1 {
		c.JSON(http.StatusOK, i.A{2, "", "", 0, 0, 0, i.A{}})
		return
	}
	matchingAdv := advs[0]
	peer := matchingAdv.NewPeer(
		u,
		q.Race,
		q.Team,
	)
	c.JSON(http.StatusOK,
		i.A{
			0,
			matchingAdv.GetIp(),
			matchingAdv.GetRelayRegion(),
			0,
			0,
			0,
			i.A{peer.Encode()},
		},
	)
}
