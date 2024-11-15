package advertisement

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
	"net"
	"net/http"
	"regexp"
)

var re *regexp.Regexp = nil

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
	if re == nil {
		// GUID Version 4
		re, _ = regexp.Compile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[89aAbB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)
	}

	game := models.G(r)
	gameTitle := game.Title()

	battleServer, battleServerExists := game.BattleServers().Get(r.PostFormValue("relayRegion"))

	if !battleServerExists {
		if !re.MatchString(r.PostFormValue("relayRegion")) {
			returnError(gameTitle, &w)
			return
		} else {
			battleServer = models.FakeBattleServer("", gameTitle != common.GameAoE1)
		}
	}

	advertisements := game.Advertisements()

	var adv shared.AdvertisementHostRequest
	if err := i.Bind(r, &adv); err == nil {
		fmt.Println(adv.MatchType)
		// Disallow Quickmatch on AoE1
		if battleServerExists && gameTitle == common.GameAoE1 && adv.Description == "SESSION_MATCH_KEY" {
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
		// Leave the previous match if the user is already in one
		// Necessary for AoE3 but might as well do it for all
		if existingAdv := u.GetAdvertisement(); existingAdv != nil {
			advertisements.RemovePeer(existingAdv, u)
		}
		storedAdv := advertisements.Store(&adv, battleServerExists && gameTitle == common.GameAoE1)
		if storedAdv == nil {
			returnError(gameTitle, &w)
			return
		}
		advertisements.NewPeer(storedAdv, u, adv.Race, adv.Team)
		response := i.A{
			0,
			storedAdv.GetId(),
			"authtoken",
		}

		var clientIp *net.IP
		if battleServerExists {
			ipStr, _, _ := net.SplitHostPort(r.RemoteAddr)
			ip := net.ParseIP(ipStr)
			clientIp = &ip
		} else {
			response = append(response, "")
		}

		response = append(
			response,
			battleServer.Encode(false, false, clientIp)...,
		)

		response = append(
			response,
			storedAdv.GetRelayRegion(),
			storedAdv.EncodePeers(),
			0,
		)

		if gameTitle == common.GameAoE2 {
			response = append(
				response,
				0,
				nil,
				nil,
				storedAdv.GetMetadata(),
				storedAdv.GetDescription(),
			)
		} else {
			response = append(response, storedAdv.GetMetadata())
		}
		i.JSON(&w, response)
	} else {
		returnError(gameTitle, &w)
	}
}
