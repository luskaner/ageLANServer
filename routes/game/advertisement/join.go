package advertisement

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/routes/game/advertisement/extra"
	"aoe2DELanServer/session"
	"github.com/gin-gonic/gin"
	"net/http"
)

type JoinRequest struct {
	extra.AdvertisementJoinRequest
	Password extra.PasswordBaseRequest
}

func Join(c *gin.Context) {
	var q JoinRequest
	if err := c.ShouldBind(&q); err != nil {
		c.JSON(http.StatusOK, j.A{2, "", "", 0, 0, 0, j.A{}})
		return
	}
	sessAny, _ := c.Get("session")
	sess, _ := sessAny.(*session.Info)
	u := sess.GetUser()
	if extra.IsInAdvertisement(u) {
		c.JSON(http.StatusOK, j.A{2, "", "", 0, 0, 0, j.A{}})
		return
	}
	advId := uint32(q.Id)
	advs := extra.FindAdvertisementsOriginal(func(adv *extra.Advertisement) bool {
		return adv.GetId() == advId && adv.GetJoinable() && adv.GetAppBinaryChecksum() == q.AppBinaryChecksum && adv.GetDataChecksum() == q.DataChecksum && adv.GetModDllFile() == q.ModDll.File && adv.GetModDllChecksum() == q.ModDll.Checksum && adv.GetModName() == q.ModName && adv.GetModVersion() == q.ModVersion && adv.GetVersionFlags() == q.VersionFlags && adv.GetPasswordValue() == q.Password.Value
	})
	if len(advs) != 1 {
		c.JSON(http.StatusOK, j.A{2, "", "", 0, 0, 0, j.A{}})
		return
	}
	matchingAdv := advs[0]
	peer := matchingAdv.NewPeer(
		u,
		q.Race,
		q.Team,
	)
	c.JSON(http.StatusOK,
		j.A{
			0,
			matchingAdv.GetIp(),
			matchingAdv.GetRelayRegion(),
			0,
			0,
			0,
			j.A{peer.Encode()},
		},
	)
}
