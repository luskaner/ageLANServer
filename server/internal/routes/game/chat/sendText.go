package chat

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/wss"
	"net/http"
	"strconv"
)

func SendText(w http.ResponseWriter, r *http.Request) {
	text := r.Form.Get("message")
	if text == "" {
		i.JSON(&w, i.A{2})
		return
	}
	chatChannelIdStr := r.Form.Get("chatroomID")
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
	sess, _ := middleware.Session(r)
	user, _ := game.Users().GetUserById(sess.GetUserId())
	if !chatChannel.HasUser(user) {
		i.JSON(&w, i.A{2})
		return
	}
	user.SendChatChannelMessage(chatChannel, text)
	i.JSON(&w, i.A{0})
	staticResponse := i.A{chatChannelIdStr, strconv.Itoa(int(user.GetId())), "", text}
	existingUsers := chatChannel.GetUsers()
	for _, existingUser := range existingUsers {
		var existingUserSession *models.Session
		existingUserSession, ok = models.GetSessionByUserId(existingUser.GetId())
		wss.SendOrStoreMessage(
			existingUserSession,
			"ChannelChatMessage",
			staticResponse,
		)
	}
}
