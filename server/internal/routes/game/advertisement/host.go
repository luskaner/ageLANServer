package advertisement

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
)

func returnError(battleServers *models.MainBattleServers, gameId string, r *http.Request, w *http.ResponseWriter) {
	battleServer := battleServers.NewBattleServer("")
	response := encodeHostResponse(
		gameId,
		2,
		0,
		battleServer,
		r,
		"",
		[]i.A{},
		"0",
		"",
	)
	i.JSON(w, response)
}

func encodeHostResponse(gameTitle string, errorCode int, advId int32, battleServer *models.MainBattleServer, r *http.Request, relayRegion string, encodedPeers []i.A, metadata string, description string) i.A {
	response := i.A{
		errorCode,
		advId,
		"authtoken",
	}
	response = append(response, battleServer.EncodeAdvertisement()...)
	response = append(
		response,
		relayRegion,
		encodedPeers,
		0,
	)
	switch gameTitle {
	case common.GameAoE1:
		response = append(response, metadata)
	case common.GameAoE2, common.GameAoM:
		response = append(
			response,
			0,
			nil,
			nil,
			metadata,
			description,
		)
	default:
		response = append(response, "0")
	}
	return response
}

func Host(w http.ResponseWriter, r *http.Request) {
	game := models.G(r)
	gameTitle := game.Title()
	battleServers := game.BattleServers()
	region := r.PostFormValue("relayRegion")
	battleServer := battleServers.NewBattleServer(region)
	if !battleServer.LAN() {
		var ok bool
		if battleServer, ok = game.BattleServers().Get(region); !ok {
			returnError(battleServers, gameTitle, r, &w)
			return
		}
	}

	var adv shared.AdvertisementHostRequest
	if err := i.Bind(r, &adv); err == nil {
		// Disallow Matchmaking as it is not implemented
		if adv.Description == "SESSION_MATCH_KEY" {
			returnError(battleServers, gameTitle, r, &w)
			return
		}
		if adv.Id != -1 {
			returnError(battleServers, gameTitle, r, &w)
			return
		}
		if gameTitle != common.GameAoE2 {
			adv.Joinable = true
		}
		u, ok := game.Users().GetUserById(adv.HostId)
		if !ok {
			returnError(battleServers, gameTitle, r, &w)
			return
		}
		advertisements := game.Advertisements()
		// Leave the previous match if the user is already in one
		// Necessary for AoE3 but might as well do it for all
		if existingAdv := advertisements.GetUserAdvertisement(u.GetId()); existingAdv != nil {
			advertisements.WithWriteLock(existingAdv.GetId(), func() {
				advertisements.UnsafeRemovePeer(existingAdv.GetId(), u.GetId())
			})
		}
		storedAdv := advertisements.Store(
			&adv,
			!battleServer.LAN(),
			gameTitle == common.GameAoM,
		)
		var response i.A
		advertisements.WithWriteLock(storedAdv.GetId(), func() {
			if advertisements.UnsafeNewPeer(storedAdv.GetId(), storedAdv.GetIp(), u.GetId(), u.GetStatId(), adv.Race, adv.Team) == nil {
				ok = false
				return
			}
			response = encodeHostResponse(
				gameTitle,
				0,
				storedAdv.GetId(),
				battleServer,
				r,
				storedAdv.GetRelayRegion(),
				storedAdv.EncodePeers(),
				storedAdv.GetMetadata(),
				storedAdv.UnsafeGetDescription(),
			)
		})
		if ok {
			i.JSON(&w, response)
		} else {
			returnError(battleServers, gameTitle, r, &w)
		}
	} else {
		returnError(battleServers, gameTitle, r, &w)
	}
}
