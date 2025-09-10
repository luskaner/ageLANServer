package advertisement

import (
	"net/http"
	"strconv"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func joinReturnStartObservingError(battleServers *models.MainBattleServers, w http.ResponseWriter) {
	battleServer := battleServers.NewBattleServer("")
	response := encodeStartObservingResponse(2, battleServer, i.A{}, i.A{}, 0)
	i.JSON(&w, response)
}

func encodeStartObservingResponse(errorCode int, battleServer *models.MainBattleServer, userIdsInt i.A, userIdsStr i.A, startTime int64) i.A {
	response := i.A{
		errorCode,
		battleServer.IPv4,
	}
	response = append(response, battleServer.EncodePorts()...)
	response = append(response, userIdsInt, startTime, userIdsStr)
	return response
}

func StartObserving(w http.ResponseWriter, r *http.Request) {
	game := models.G(r)
	battleServers := game.BattleServers()
	var q searchQuery
	if err := i.Bind(r, &q); err != nil {
		joinReturnStartObservingError(battleServers, w)
		return
	}
	advId := r.PostFormValue("advertisementid")
	advIdInt64, err := strconv.ParseInt(advId, 10, 32)
	if err != nil {
		joinReturnStartObservingError(battleServers, w)
		return
	}
	advIdInt := int32(advIdInt64)
	sess := middleware.Session(r)
	currentUserId := sess.GetUserId()
	advertisements := game.Advertisements()

	adv, foundAdv := advertisements.GetAdvertisement(advIdInt)
	if !foundAdv {
		joinReturnStartObservingError(battleServers, w)
		return
	}
	advertisements.WithReadLock(adv.GetId(), func() {
		_, matchesBattleServer := battleServers.Get(adv.GetRelayRegion())
		peers := adv.GetPeers()
		_, isPeer := peers.Load(currentUserId)
		if adv.UnsafeGetJoinable() &&
			isPeer &&
			adv.UnsafeGetVisible() &&
			adv.UnsafeGetAppBinaryChecksum() != q.AppBinaryChecksum &&
			adv.UnsafeGetDataChecksum() != q.DataChecksum &&
			adv.UnsafeGetMatchType() != q.MatchType &&
			adv.UnsafeGetModDllFile() != q.ModDllFile &&
			adv.UnsafeGetModDllChecksum() != q.ModDllChecksum &&
			adv.UnsafeGetModName() != q.ModName &&
			adv.UnsafeGetModVersion() != q.ModVersion &&
			adv.UnsafeGetVersionFlags() != q.VersionFlags &&
			!adv.UnsafeGetObserversEnabled() &&
			!matchesBattleServer {
			joinReturnStartObservingError(battleServers, w)
			return
		}
	})
	var response i.A
	advertisements.WithReadLock(adv.GetId(), func() {
		battleServer, battleServerExists := game.BattleServers().Get(adv.GetRelayRegion())
		if !battleServerExists {
			joinReturnStartObservingError(battleServers, w)
			return
		}
		peers := adv.GetPeers()
		peerLength, userIds := peers.Keys()
		userIdsInt := make(i.A, peerLength)
		userIdsStr := make(i.A, peerLength)
		j := 0
		for userId := range userIds {
			userIdsInt[j] = i.A{userId, i.A{}}
			userIdsStr[j] = i.A{strconv.Itoa(int(userId)), i.A{}}
			j++
		}
		response = encodeStartObservingResponse(
			0,
			battleServer,
			userIdsInt,
			userIdsStr,
			adv.UnsafeGetStartTime(),
		)
	})
	adv.StartObserving(sess.GetUserId())
	i.JSON(&w, response)
}
