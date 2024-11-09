package advertisement

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
	"net/http"
)

type JoinRequest struct {
	shared.AdvertisementBaseRequest
	Password string `schema:"password"`
}

func joinReturnError(w http.ResponseWriter) {
	i.JSON(&w, i.A{2, "", "", 0, 0, 0, i.A{}})
}

func Join(w http.ResponseWriter, r *http.Request) {
	var q JoinRequest
	if err := i.Bind(r, &q); err != nil {
		joinReturnError(w)
		return
	}
	sess, _ := middleware.Session(r)
	game := models.G(r)
	u, _ := game.Users().GetUserById(sess.GetUserId())
	advertisements := game.Advertisements()
	// Leave the previous match if the user is already in one
	// Necessary for AoE1 but might as well do it for all
	if existingAdv := u.GetAdvertisement(); existingAdv != nil {
		advertisements.RemovePeer(existingAdv, u)
	}

	advs := advertisements.FindAdvertisements(func(adv *models.MainAdvertisement) bool {
		return adv.GetId() == q.Id &&
			adv.GetJoinable() &&
			adv.GetAppBinaryChecksum() == q.AppBinaryChecksum &&
			adv.GetDataChecksum() == q.DataChecksum &&
			adv.GetModDllFile() == q.ModDllFile &&
			adv.GetModDllChecksum() == q.ModDllChecksum &&
			adv.GetModName() == q.ModName &&
			adv.GetModVersion() == q.ModVersion &&
			adv.GetVersionFlags() == q.VersionFlags &&
			adv.GetPasswordValue() == q.Password
	})
	if len(advs) != 1 {
		joinReturnError(w)
		return
	}
	matchingAdv := advs[0]
	peer := advertisements.NewPeer(
		matchingAdv,
		u,
		q.Race,
		q.Team,
	)
	var relayRegion string
	gameTitle := game.Title()
	if gameTitle == common.GameAoE2 {
		relayRegion = matchingAdv.GetRelayRegion()
	}
	response := i.A{
		0,
		matchingAdv.GetIp(),
		relayRegion,
		0,
		0,
	}
	if gameTitle != common.GameAoE1 {
		response = append(response, 0)
	}
	response = append(response, i.A{peer.Encode()})
	i.JSON(&w, response)
}
