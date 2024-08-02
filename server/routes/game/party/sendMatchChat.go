package party

import (
	"encoding/json"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/middleware"
	"github.com/luskaner/aoe2DELanServer/server/models"
	"github.com/luskaner/aoe2DELanServer/server/routes/wss"
	"net/http"
)

type request struct {
	ToProfileIdsStr string `schema:"to_profile_ids"`
	MessageTypeID   uint8  `schema:"messageTypeID"`
	MatchID         int32  `schema:"match_id"`
	Broadcast       bool   `schema:"broadcast"`
	Message         string `schema:"message"`
}

func SendMatchChat(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := i.Bind(r, &req); err != nil {
		i.JSON(&w, i.A{2})
		return
	}

	var toProfileIds []int32
	err := json.Unmarshal([]byte(req.ToProfileIdsStr), &toProfileIds)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}

	adv, ok := models.GetAdvertisement(req.MatchID)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}

	sess, _ := middleware.Session(r)
	currentUser := sess.GetUser()

	// Only peers within the match can send messages
	// What about AI?
	if _, ok := adv.GetPeer(currentUser); !ok {
		i.JSON(&w, i.A{2})
		return
	}

	receivers := make([]*models.User, len(toProfileIds))
	for j, profileId := range toProfileIds {
		receivers[j], _ = models.GetUserById(profileId)
	}

	message := adv.AddMessage(
		req.Broadcast,
		req.Message,
		req.MessageTypeID,
		currentUser,
		receivers,
	)

	messageEncoded := message.Encode()
	for _, receiver := range receivers {
		receiverSession, ok := models.GetSessionByUser(receiver)
		if !ok {
			continue
		}
		go func() {
			wss.SendMessage(
				receiverSession.GetId(),
				i.A{
					0,
					"MatchReceivedChatMessage",
					receiver.GetId(),
					messageEncoded,
				},
			)
		}()
	}
	i.JSON(&w, i.A{0})
}
