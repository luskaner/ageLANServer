package advertisement

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/advertisement/shared"
	"net/http"
)

type JoinRequest struct {
	shared.AdvertisementBaseRequest
	Password string `schema:"password"`
}

func Join(w http.ResponseWriter, r *http.Request) {
	var q JoinRequest
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, i.A{2, "", "", 0, 0, 0, i.A{}})
		return
	}
	sess, _ := middleware.Session(r)
	game := models.G(r)
	u, _ := game.Users().GetUserById(sess.GetUserId())
	advertisements := game.Advertisements()
	if advertisements.IsInAdvertisement(u) {
		i.JSON(&w, i.A{2, "", "", 0, 0, 0, i.A{}})
		return
	}
	advs := advertisements.FindAdvertisements(func(adv *models.MainAdvertisement) bool {
		return adv.GetId() == q.Id && adv.GetJoinable() && adv.GetAppBinaryChecksum() == q.AppBinaryChecksum && adv.GetDataChecksum() == q.DataChecksum && adv.GetModDllFile() == q.ModDllFile && adv.GetModDllChecksum() == q.ModDllChecksum && adv.GetModName() == q.ModName && adv.GetModVersion() == q.ModVersion && adv.GetVersionFlags() == q.VersionFlags && adv.GetPasswordValue() == q.Password
	})
	if len(advs) != 1 {
		i.JSON(&w, i.A{2, "", "", 0, 0, 0, i.A{}})
		return
	}
	matchingAdv := advs[0]
	peer := advertisements.NewPeer(
		matchingAdv,
		u,
		q.Race,
		q.Team,
	)
	i.JSON(&w,
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
