package party

import (
	"net/http"
	"slices"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

type request struct {
	MessageTypeID uint8  `schema:"messageTypeID"`
	MatchID       int32  `schema:"match_id"`
	Broadcast     bool   `schema:"broadcast"`
	Message       string `schema:"message"`
}

type profileIds struct {
	Ids i.Json[[]int32] `schema:"to_profile_ids"`
}

type profileId struct {
	Id int32 `schema:"to_profile_id"`
}

func SendMatchChat(w http.ResponseWriter, r *http.Request) {
	// FIXME: AoE3 duplicate messages/wrong total
	var req request
	if err := i.Bind(r, &req); err != nil {
		i.JSON(&w, i.A{2})
		return
	}

	var toProfileIds profileIds
	game := models.G(r)
	if game.Title() == common.GameAoE3 {
		var toProfileId profileId
		if err := i.Bind(r, &toProfileId); err != nil {
			i.JSON(&w, i.A{2})
			return
		}
		toProfileIds.Ids.Data = append(toProfileIds.Ids.Data, toProfileId.Id)
	} else if err := i.Bind(r, &toProfileIds); err != nil {
		i.JSON(&w, i.A{2})
		return
	}

	adv, ok := game.Advertisements().GetAdvertisement(req.MatchID)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}

	sess := models.SessionOrPanic(r)
	currentUserId := sess.GetUserId()
	peers := adv.GetPeers()

	// Only peers within the match can send messages
	if _, ok = peers.Load(currentUserId); !ok {
		i.JSON(&w, i.A{2})
		return
	}

	users := game.Users()
	if game.Title() == common.GameAoM {
		toProfileIds.Ids.Data = slices.DeleteFunc(toProfileIds.Ids.Data, func(id int32) bool { return id == currentUserId })
	}
	receivers := make([]models.User, len(toProfileIds.Ids.Data))

	for j, profId := range toProfileIds.Ids.Data {
		receivers[j], ok = users.GetUserById(profId)
		if !ok {
			i.JSON(&w, i.A{2})
			return
		}
	}

	currentUser, ok := game.Users().GetUserById(currentUserId)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	message := adv.MakeMessage(
		req.Broadcast,
		req.Message,
		req.MessageTypeID,
		currentUser,
		receivers,
	)

	messageEncoded := message.Encode()
	sessions := game.Sessions()
	var receiverSession models.Session
	for _, receiver := range receivers {
		receiverSession, ok = sessions.GetByUserId(receiver.GetId())
		if !ok {
			continue
		}
		wss.SendOrStoreMessage(
			receiverSession,
			"MatchReceivedChatMessage",
			messageEncoded,
		)
	}
	i.JSON(&w, i.A{0})
}
