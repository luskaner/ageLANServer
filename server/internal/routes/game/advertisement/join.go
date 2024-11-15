package advertisement

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
	"net"
	"net/http"
)

type JoinRequest struct {
	shared.AdvertisementBaseRequest
	Password string `schema:"password"`
}

func joinReturnError(gameTitle string, w http.ResponseWriter) {
	response := i.A{2,
		"",
		"",
		0,
		0,
	}
	if gameTitle != common.GameAoE1 {
		response = append(response, 0)
	}
	response = append(response, i.A{0})
	i.JSON(&w, response)
}

func Join(w http.ResponseWriter, r *http.Request) {
	game := models.G(r)
	gameTitle := game.Title()
	var q JoinRequest
	if err := i.Bind(r, &q); err != nil {
		joinReturnError(gameTitle, w)
		return
	}
	sess, _ := middleware.Session(r)

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
		joinReturnError(gameTitle, w)
		return
	}
	matchingAdv := advs[0]
	peer := advertisements.NewPeer(
		matchingAdv,
		u,
		q.Race,
		q.Team,
	)

	var battleServer *models.MainBattleServer
	var clientIP *net.IP
	var ok bool
	if battleServer, ok = game.BattleServers().Get(matchingAdv.GetRelayRegion()); !ok {
		var serverIP string
		if gameTitle == common.GameAoE2 {
			serverIP = ""
		} else {
			serverIP = matchingAdv.GetRelayRegion()
		}
		battleServer = models.FakeBattleServer(serverIP, gameTitle != common.GameAoE1)
	} else {
		ipStr, _, _ := net.SplitHostPort(r.RemoteAddr)
		IP := net.ParseIP(ipStr)
		clientIP = &IP
	}
	response := i.A{0, matchingAdv.GetIp()}
	if !ok {
		response = append(response, "")
	}
	response = append(response, battleServer.Encode(false, false, clientIP)...)
	response = append(response, i.A{peer.Encode()})
	i.JSON(&w, response)
}
