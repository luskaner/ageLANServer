package advertisement

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
)

func returnError(battleServers models.BattleServers, gameId string, r *http.Request, w *http.ResponseWriter) {
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

func encodeHostResponse(gameTitle string, errorCode int, advId int32, battleServer models.BattleServer, r *http.Request, relayRegion string, encodedPeers []i.A, metadata string, description string) i.A {
	response := i.A{
		errorCode,
		advId,
		"authtoken",
	}
	response = append(response, battleServer.EncodeAdvertisement(r)...)
	response = append(
		response,
		relayRegion,
		encodedPeers,
		0,
	)
	switch gameTitle {
	case common.GameAoE1:
		response = append(response, metadata)
	case common.GameAoE2, common.GameAoE4, common.GameAoM:
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
		// In AoE4 we cannot differentiate between Matchmaking and custom matches so just allow it
		if gameTitle != common.GameAoE4 && adv.Description == "SESSION_MATCH_KEY" {
			// Disallow Matchmaking as it is not implemented
			returnError(battleServers, gameTitle, r, &w)
			return
		}
		if adv.Id != -1 {
			returnError(battleServers, gameTitle, r, &w)
			return
		}
		if gameTitle == common.GameAoE1 || gameTitle == common.GameAoE3 || gameTitle == common.GameAoM {
			adv.Joinable = true
		}
		u, ok := game.Users().GetUserById(adv.HostId)
		if !ok {
			returnError(battleServers, gameTitle, r, &w)
			return
		}
		advertisements := game.Advertisements()
		// Leave the previous match if the user is already in one
		// Necessary for AoE3 but might as well do it for all (except AoE4 which needs multiple for groups)
		if gameTitle != common.GameAoE4 {
			// FIXME: Exit in aoe4 if the currrent match is not a party
			if existingAdv := advertisements.GetUserAdvertisement(u.GetId()); existingAdv != nil {
				advertisements.WithWriteLock(existingAdv.GetId(), func() {
					advertisements.UnsafeRemovePeer(existingAdv.GetId(), u.GetId())
				})
			}
		}
		if adv.Party != -1 {
			if partyAdv, ok := advertisements.GetAdvertisement(adv.Party); !ok {
				returnError(battleServers, gameTitle, r, &w)
			} else if partyAdv.GetParty() != -1 {
				returnError(battleServers, gameTitle, r, &w)
			}
		}
		storedAdv := advertisements.Store(
			&adv,
			!battleServer.LAN(),
			gameTitle,
		)
		var response i.A
		advertisements.WithWriteLock(storedAdv.GetId(), func() {
			if advertisements.UnsafeNewPeer(storedAdv.GetId(), storedAdv.GetIp(), u.GetId(), u.GetStatId(), adv.Party, adv.Race, adv.Team) == nil {
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
				storedAdv.GetXboxSessionId(),
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
