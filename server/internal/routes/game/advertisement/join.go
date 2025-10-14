package advertisement

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
)

type JoinRequest struct {
	shared.AdvertisementBaseRequest
	Password string `schema:"password"`
}

func joinReturnError(battleServers *models.MainBattleServers, r *http.Request, w http.ResponseWriter) {
	battleServer := battleServers.NewBattleServer("")
	response := encodeJoinResponse(2, "", battleServer, r, i.A{})
	i.JSON(&w, response)
}

func encodeJoinResponse(errorCode int, ip string, battleServer *models.MainBattleServer, r *http.Request, peerEncoded i.A) i.A {
	response := i.A{
		errorCode,
		ip,
		battleServer.IPv4,
	}
	response = append(response, battleServer.EncodePorts()...)
	response = append(response, i.A{peerEncoded})
	return response
}

func Join(w http.ResponseWriter, r *http.Request) {
	game := models.G(r)
	battleServers := game.BattleServers()
	var q JoinRequest
	if err := i.Bind(r, &q); err != nil {
		joinReturnError(battleServers, r, w)
		return
	}
	sess := middleware.SessionOrPanic(r)

	u, ok := game.Users().GetUserById(sess.GetUserId())
	if !ok {
		joinReturnError(battleServers, r, w)
		return
	}
	advertisements := game.Advertisements()
	// Leave the previous match if the user is already in one
	// Necessary for AoE1 but might as well do it for all
	if existingAdv := advertisements.GetUserAdvertisement(u.GetId()); existingAdv != nil {
		advertisements.WithWriteLock(existingAdv.GetId(), func() {
			advertisements.UnsafeRemovePeer(existingAdv.GetId(), u.GetId())
		})
	}
	matchingAdv, foundAdv := advertisements.GetAdvertisement(q.Id)
	if !foundAdv {
		joinReturnError(battleServers, r, w)
		return
	}
	advertisements.WithReadLock(matchingAdv.GetId(), func() {
		if !matchingAdv.UnsafeGetJoinable() ||
			matchingAdv.UnsafeGetAppBinaryChecksum() != q.AppBinaryChecksum ||
			matchingAdv.UnsafeGetDataChecksum() != q.DataChecksum ||
			matchingAdv.UnsafeGetModDllFile() != q.ModDllFile ||
			matchingAdv.UnsafeGetModDllChecksum() != q.ModDllChecksum ||
			matchingAdv.UnsafeGetModName() != q.ModName ||
			matchingAdv.UnsafeGetModVersion() != q.ModVersion ||
			matchingAdv.UnsafeGetVersionFlags() != q.VersionFlags ||
			matchingAdv.UnsafeGetPasswordValue() != q.Password {
			joinReturnError(battleServers, r, w)
			return
		}
	})
	var response i.A
	advertisements.WithWriteLock(matchingAdv.GetId(), func() {
		peer := advertisements.UnsafeNewPeer(
			matchingAdv.GetId(),
			matchingAdv.GetIp(),
			u.GetId(),
			u.GetStatId(),
			q.Race,
			q.Team,
		)
		if peer == nil {
			ok = false
			return
		}
		battleServer, battleServerExists := battleServers.Get(matchingAdv.GetRelayRegion())
		if !battleServerExists {
			battleServer = battleServers.NewLANBattleServer("")
		}
		response = encodeJoinResponse(
			0,
			matchingAdv.GetIp(),
			battleServer,
			r,
			peer.Encode(),
		)
	})
	if ok {
		i.JSON(&w, response)
	} else {
		joinReturnError(battleServers, r, w)
	}
}
