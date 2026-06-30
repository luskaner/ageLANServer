package advertisement

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common/game"
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
	case game.AoE1:
		response = append(response, metadata)
	case game.AoE2, game.AoE4, game.AoM:
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
	g := models.G(r)
	gameTitle := g.Title()
	battleServers := g.BattleServers()
	region := r.PostFormValue("relayRegion")
	battleServer := battleServers.NewBattleServer(region)
	if !battleServer.LAN() {
		var ok bool
		if battleServer, ok = g.BattleServers().Get(region); !ok {
			returnError(battleServers, gameTitle, r, &w)
			return
		}
	}

	var adv shared.AdvertisementHostRequest
	if err := i.Bind(r, &adv); err == nil {
		// In AoE4 we cannot differentiate between Matchmaking and custom matches so just allow it
		if gameTitle != game.AoE4 && adv.Description == "SESSION_MATCH_KEY" {
			// Disallow Matchmaking as it is not implemented
			returnError(battleServers, gameTitle, r, &w)
			return
		}
		if adv.Id != -1 {
			returnError(battleServers, gameTitle, r, &w)
			return
		}
		if gameTitle == game.AoE1 || gameTitle == game.AoE3 || gameTitle == game.AoM {
			adv.Joinable = true
		}
		u, ok := g.Users().GetUserById(adv.HostId)
		if !ok {
			returnError(battleServers, gameTitle, r, &w)
			return
		}
		advertisements := g.Advertisements()
		// Leave the previous match if the user is already in one
		// Necessary for AoE3 but might as well do it for all (except AoE4 which needs multiple for groups)
		if gameTitle != game.AoE4 {
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
