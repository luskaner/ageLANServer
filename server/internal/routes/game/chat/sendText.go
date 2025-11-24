package chat

import (
	"net/http"
	"strconv"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

func SendText(w http.ResponseWriter, r *http.Request) {
	text := r.FormValue("message")
	if text == "" {
		i.JSON(&w, i.A{2})
		return
	}
	chatChannelIdStr := r.FormValue("chatroomID")
	if chatChannelIdStr == "" {
		i.JSON(&w, i.A{2})
		return
	}
	chatChannelId, err := strconv.ParseInt(chatChannelIdStr, 10, 32)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	game := models.G(r)
	chatChannel, ok := game.ChatChannels().GetById(int32(chatChannelId))
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
	staticResponse := i.A{chatChannelIdStr, strconv.Itoa(int(user.GetId())), "", text}
	for existingUser := range chatChannel.GetUsers() {
		var existingUserSession *models.Session
		existingUserSession, ok = models.GetSessionByUserId(existingUser.GetId())
		wss.SendOrStoreMessage(
			existingUserSession,
			"ChannelChatMessage",
			staticResponse,
		)
	}
}
