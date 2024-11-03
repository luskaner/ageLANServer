package chat

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
	"net/http"
	"strconv"
)

func SendWhisper(w http.ResponseWriter, r *http.Request) {
	// FIXME: Show people as offline always
	text := r.Form.Get("message")
	if text == "" {
		i.JSON(&w, i.A{2})
		return
	}
	targetUserIdStr := r.Form.Get("recipientID")
	if targetUserIdStr == "" {
		i.JSON(&w, i.A{2})
		return
	}
	targetUserId, err := strconv.ParseInt(targetUserIdStr, 10, 32)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	session, ok := models.GetSessionByUserId(int32(targetUserId))
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	currentSession, _ := middleware.Session(r)
	currentUser, _ := models.G(r).Users().GetUserById(currentSession.GetUserId())
	i.JSON(&w, i.A{0})
	wss.SendOrStoreMessage(
		session,
		"PersonalChatMessage",
		i.A{
			currentUser.GetProfileInfo(false),
			"",
			text,
		},
	)
}
