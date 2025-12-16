package chat

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

type recipientIDs struct {
	IDs i.Json[[]int32] `schema:"recipientIDs"`
}

type recipientID struct {
	ID int32 `schema:"recipientID"`
}

func whisperResult(w *http.ResponseWriter, gameId string, code int) {
	response := i.A{code}
	if gameId == common.GameAoM {
		response = append(response, i.A{0})
	}
	i.JSON(w, response)
}

func SendWhisper(w http.ResponseWriter, r *http.Request) {
	game := models.G(r)
	gameTitle := game.Title()
	var req textRequest
	err := i.Bind(r, &req)
	if err != nil {
		whisperResult(&w, gameTitle, 2)
		return
	}

	var targetUserIds recipientIDs
	if gameTitle == common.GameAoM {
		if err := i.Bind(r, &targetUserIds); err != nil {
			whisperResult(&w, gameTitle, 2)
			return
		}
	} else {
		var recpId recipientID
		if err = i.Bind(r, &recpId); err != nil {
			whisperResult(&w, gameTitle, 2)
			return
		}
		targetUserIds.IDs.Data = append(targetUserIds.IDs.Data, recpId.ID)
	}
	users := game.Users()
	receivers := make([]models.User, len(targetUserIds.IDs.Data))
	var ok bool

	for j, profileId := range targetUserIds.IDs.Data {
		receivers[j], ok = users.GetUserById(profileId)
		if !ok {
			whisperResult(&w, gameTitle, 2)
			return
		}
	}
	currentSession := models.SessionOrPanic(r)
	currentUser, ok := game.Users().GetUserById(currentSession.GetUserId())
	if !ok {
		whisperResult(&w, gameTitle, 2)
		return
	}

	message := i.A{""}
	if gameTitle == common.GameAoM {
		message = append(
			message,
			i.A{
				"",
				req.Message,
			}...,
		)
	}
	message = append(message, req.Message)
	sessions := game.Sessions()
	var receiverSession models.Session
	for _, receiver := range receivers {
		receiverSession, ok = sessions.GetByUserId(receiver.GetId())
		if !ok {
			continue
		}
		finalMessage := i.A{
			currentUser.GetProfileInfo(false, receiverSession.GetClientLibVersion()),
		}
		finalMessage = append(finalMessage, message...)
		wss.SendOrStoreMessage(
			receiverSession,
			"PersonalChatMessage",
			finalMessage,
		)
	}

	whisperResult(&w, gameTitle, 0)
}
