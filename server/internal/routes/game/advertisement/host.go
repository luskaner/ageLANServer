package advertisement

import (
	"github.com/luskaner/aoe2DELanServer/common"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/advertisement/shared"
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
		if advertisements.IsInAdvertisement(u) {
			if gameTitle == common.GameAoE3 {
				// TODO: Leave advertisement to enable Rematch
			} else {
				returnError(&w)
				return
			}
		}
		storedAdv := advertisements.Store(&adv)
		if storedAdv == nil {
			returnError(&w)
			return
		}
		advertisements.NewPeer(storedAdv, u, adv.Race, adv.Team)

		if gameTitle == common.GameAoE3 {
			i.JSON(&w,
				i.A{
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
					"0",
				},
			)
		} else {
			i.JSON(&w,
				i.A{
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
					0,
					nil,
					nil,
					"0",
					storedAdv.GetDescription(),
				},
			)
		}
	} else {
		returnError(&w)
	}
}
