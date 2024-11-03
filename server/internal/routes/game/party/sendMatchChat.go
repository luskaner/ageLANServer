package party

import (
	"encoding/json"
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
	"net/http"
	"strconv"
)

type request struct {
	MessageTypeID uint8  `schema:"messageTypeID"`
	MatchID       int32  `schema:"match_id"`
	Broadcast     bool   `schema:"broadcast"`
	Message       string `schema:"message"`
}

func SendMatchChat(w http.ResponseWriter, r *http.Request) {
	// FIXME: AoE3 duplicate messages/wrong total
	var req request
	if err := i.Bind(r, &req); err != nil {
		i.JSON(&w, i.A{2})
		return
	}

	var toProfileIds []int32
	game := models.G(r)
	if game.Title() == common.GameAoE3 {
		profileIdStr := r.FormValue("to_profile_id")
		if profileIdStr == "" {
			i.JSON(&w, i.A{0})
			return
		}
		profileId, err := strconv.ParseInt(profileIdStr, 10, 32)
		if err != nil {
			i.JSON(&w, i.A{2})
			return
		}
		toProfileIds = append(toProfileIds, int32(profileId))
	} else {
		err := json.Unmarshal([]byte(r.FormValue("to_profile_ids")), &toProfileIds)
		if err != nil {
			i.JSON(&w, i.A{2})
			return
		}
	}

	adv, ok := game.Advertisements().GetAdvertisement(req.MatchID)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}

	sess, _ := middleware.Session(r)
	currentUser, _ := game.Users().GetUserById(sess.GetUserId())

	// Only peers within the match can send messages
	// What about AI?
	if _, ok = adv.GetPeer(currentUser); !ok {
		i.JSON(&w, i.A{2})
		return
	}

	users := game.Users()

	receivers := make([]*models.MainUser, len(toProfileIds))
	for j, profileId := range toProfileIds {
		receivers[j], _ = users.GetUserById(profileId)
	}

	message := adv.AddMessage(
		req.Broadcast,
		req.Message,
		req.MessageTypeID,
		currentUser,
		receivers,
	)

	messageEncoded := message.Encode()
	var receiverSession *models.Session
	for _, receiver := range receivers {
		receiverSession, ok = models.GetSessionByUserId(receiver.GetId())
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
