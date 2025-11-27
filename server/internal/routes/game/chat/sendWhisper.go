package chat

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

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
	text := r.FormValue("message")
	if text == "" {
		whisperResult(&w, gameTitle, 2)
		return
	}

	var targetUserIds []int32
	if gameTitle == common.GameAoM {
		targetUserIdsStr := r.FormValue("recipientIDs")
		if targetUserIdsStr == "" {
			whisperResult(&w, gameTitle, 2)
			return
		}
		err := json.Unmarshal([]byte(targetUserIdsStr), &targetUserIds)
		if err != nil {
			whisperResult(&w, gameTitle, 2)
			return
		}
	} else {
		targetUserIdStr := r.FormValue("recipientID")
		if targetUserIdStr == "" {
			whisperResult(&w, gameTitle, 2)
			return
		}
		targetUserId, err := strconv.ParseInt(targetUserIdStr, 10, 32)
		if err != nil {
			whisperResult(&w, gameTitle, 2)
			return
		}
		targetUserIds = append(targetUserIds, int32(targetUserId))
	}
	users := game.Users()
	receivers := make([]models.User, len(targetUserIds))
	var ok bool

	for j, profileId := range targetUserIds {
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
				text,
			}...,
		)
	}
	message = append(message, text)
	var receiverSession *models.Session
	for _, receiver := range receivers {
		receiverSession, ok = models.GetSessionByUserId(receiver.GetId())
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
