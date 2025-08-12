package advertisement

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
	"net/http"
)

type JoinRequest struct {
	shared.AdvertisementBaseRequest
	Password string `schema:"password"`
}

func joinReturnError(gameTitle common.GameTitle, w http.ResponseWriter) {
	response := i.A{2,
		"",
		"",
		0,
		0,
	}
	if gameTitle != common.AoE1 {
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
	sess := middleware.Session(r)

	u, ok := game.Users().GetUserById(sess.GetUserId())
	if !ok {
		joinReturnError(gameTitle, w)
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

	matchingAdv := advertisements.UnsafeFirstAdvertisement(func(adv *models.MainAdvertisement) bool {
		var matches bool
		advertisements.WithReadLock(adv.GetId(), func() {
			matches = adv.GetId() == q.Id &&
				adv.UnsafeGetJoinable() &&
				adv.UnsafeGetAppBinaryChecksum() == q.AppBinaryChecksum &&
				adv.UnsafeGetDataChecksum() == q.DataChecksum &&
				adv.UnsafeGetModDllFile() == q.ModDllFile &&
				adv.UnsafeGetModDllChecksum() == q.ModDllChecksum &&
				adv.UnsafeGetModName() == q.ModName &&
				adv.UnsafeGetModVersion() == q.ModVersion &&
				adv.UnsafeGetVersionFlags() == q.VersionFlags &&
				adv.UnsafeGetPasswordValue() == q.Password
		})
		return matches
	})
	if matchingAdv == nil {
		joinReturnError(gameTitle, w)
		return
	}
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
		var relayRegion string
		if gameTitle == common.AoE2 {
			relayRegion = matchingAdv.GetRelayRegion()
		}
		response = i.A{
			0,
			matchingAdv.GetIp(),
			relayRegion,
			0,
			0,
		}
		if gameTitle != common.AoE1 {
			response = append(response, 0)
		}
		response = append(response, i.A{peer.Encode()})
	})
	if ok {
		i.JSON(&w, response)
	} else {
		joinReturnError(gameTitle, w)
	}
}
