package chat

import (
	"net/http"
	"strconv"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

type textRequest struct {
	Message string `schema:"message"`
}

type sendTextRequest struct {
	chatroomRequest
	textRequest
}

func SendText(w http.ResponseWriter, r *http.Request) {
	var req sendTextRequest
	err := i.Bind(r, &req)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	game := models.G(r)
	chatChannel, ok := game.ChatChannels().GetById(req.ChatroomID)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	sess := models.SessionOrPanic(r)
	user, ok := game.Users().GetUserById(sess.GetUserId())
	if !ok || !chatChannel.HasUser(user) {
		i.JSON(&w, i.A{2})
		return
	}
	i.JSON(&w, i.A{0})
	sessions := game.Sessions()
	staticResponse := i.A{strconv.Itoa(int(req.ChatroomID)), strconv.Itoa(int(user.GetId())), "", req.Message}
	for existingUser := range chatChannel.GetUsers() {
		var existingUserSession models.Session
		existingUserSession, ok = sessions.GetByUserId(existingUser.GetId())
		wss.SendOrStoreMessage(
			existingUserSession,
			"ChannelChatMessage",
			staticResponse,
		)
	}
}
