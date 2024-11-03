package advertisement

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
	"net/http"
	"regexp"
)

var re *regexp.Regexp = nil

func returnError(w *http.ResponseWriter) {
	i.JSON(w, i.A{
		2,
		0,
		"",
		"",
		0,
		0,
		0,
		"",
		i.A{},
		0,
		0,
		nil,
		nil,
		"",
		"",
	})
}

func Host(w http.ResponseWriter, r *http.Request) {
	if re == nil {
		// GUID Version 4
		re, _ = regexp.Compile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[89aAbB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)
	}

	// Only LAN servers are allowed and need the GUID to store it
	if !re.MatchString(r.PostFormValue("relayRegion")) {
		returnError(&w)
		return
	}

	game := models.G(r)
	gameTitle := game.Title()
	advertisements := game.Advertisements()

	var adv shared.AdvertisementHostRequest
	if err := i.Bind(r, &adv); err == nil {
		if gameTitle == common.GameAoE3 {
			adv.Joinable = true
		}
		u, ok := game.Users().GetUserById(adv.HostId)
		if !ok {
			returnError(&w)
			return
		}
		// Leave the previous match if the user is already in one
		// Necessary for AoE3 but might as well do it for all
		if existingAdv := u.GetAdvertisement(); existingAdv != nil {
			advertisements.RemovePeer(existingAdv, u)
		}
		storedAdv := advertisements.Store(&adv)
		if storedAdv == nil {
			returnError(&w)
			return
		}
		advertisements.NewPeer(storedAdv, u, adv.Race, adv.Team)
		response := i.A{
			0,
			storedAdv.GetId(),
			"authtoken",
			"",
			0,
			0,
			0,
			storedAdv.GetRelayRegion(),
			storedAdv.EncodePeers(),
			0,
		}
		if gameTitle == common.GameAoE3 {
			response = append(response, "0")
		}
		if gameTitle == common.GameAoE2 {
			response = append(
				response,
				0,
				nil,
				nil,
				"0",
				storedAdv.GetDescription(),
			)
		}
		i.JSON(&w, response)
	} else {
		returnError(&w)
	}
}
