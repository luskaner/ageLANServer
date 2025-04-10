package advertisement

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
	"net/http"
	"regexp"
)

// GUID Version 4
var re, _ = regexp.Compile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[89aAbB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)

func returnError(gameId string, w *http.ResponseWriter) {
	response := i.A{
		2,
		0,
		"authtoken",
		"",
		0,
		0,
	}
	if gameId != common.GameAoE1 {
		response = append(response, 0)
	}

	response = append(
		response,
		"",
		i.A{},
		0,
	)

	if gameId != common.GameAoE2 {
		response = append(response, "0")
	}
	if gameId == common.GameAoE2 {
		response = append(
			response,
			0,
			nil,
			nil,
			"0",
			"",
		)
	}
	i.JSON(w, response)
}

func Host(w http.ResponseWriter, r *http.Request) {
	game := models.G(r)
	gameTitle := game.Title()

	// Only LAN servers are allowed and need the GUID to store it
	if !re.MatchString(r.PostFormValue("relayRegion")) {
		returnError(gameTitle, &w)
		return
	}

	var adv shared.AdvertisementHostRequest
	if err := i.Bind(r, &adv); err == nil {
		if adv.Id != -1 {
			returnError(gameTitle, &w)
			return
		}
		if gameTitle != common.GameAoE2 {
			adv.Joinable = true
		}
		u, ok := game.Users().GetUserById(adv.HostId)
		if !ok {
			returnError(gameTitle, &w)
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
		storedAdv := advertisements.Store(&adv)
		var response i.A
		advertisements.WithWriteLock(storedAdv.GetId(), func() {
			if advertisements.UnsafeNewPeer(storedAdv.GetId(), storedAdv.GetIp(), u.GetId(), u.GetStatId(), adv.Race, adv.Team) == nil {
				ok = false
				return
			}
			response = i.A{
				0,
				storedAdv.GetId(),
				"authtoken",
				"",
				0,
				0,
			}
			if gameTitle != common.GameAoE1 {
				response = append(response, 0)
			}

			response = append(
				response,
				storedAdv.GetRelayRegion(),
				storedAdv.EncodePeers(),
				0,
			)

			if gameTitle != common.GameAoE2 {
				response = append(response, "0")
			}
			if gameTitle == common.GameAoE2 {
				response = append(
					response,
					0,
					nil,
					nil,
					"0",
					storedAdv.UnsafeGetDescription(),
				)
			}
		})
		if ok {
			i.JSON(&w, response)
		} else {
			returnError(gameTitle, &w)
		}
	} else {
		returnError(gameTitle, &w)
	}
}
