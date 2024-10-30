package chat

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/wss"
	"net/http"
	"strconv"
)

func JoinChannel(w http.ResponseWriter, r *http.Request) {
	chatChannelIdStr := r.Form.Get("chatroomID")
	if chatChannelIdStr == "" {
		i.JSON(&w, i.A{2, "", 0, i.A{}})
		return
	}
	chatChannelId, err := strconv.Atoi(chatChannelIdStr)
	if err != nil {
		i.JSON(&w, i.A{2, "", 0, i.A{}})
		return
	}
	game := models.G(r)
	chatChannel, ok := game.ChatChannels().GetById(int32(chatChannelId))
	if !ok {
		i.JSON(&w, i.A{2, "", 0, i.A{}})
		return
	}
	sess, _ := middleware.Session(r)
	user, _ := game.Users().GetUserById(sess.GetUserId())
	if chatChannel.HasUser(user) {
		i.JSON(&w, i.A{2, "", 0, i.A{}})
		return
	}
	i.JSON(&w, i.A{0, chatChannelIdStr, 0, chatChannel.EncodeUsers()})
	user.JoinChatChannel(chatChannel)
	staticResponse := i.A{chatChannelIdStr, i.A{0, user.GetProfileInfo(false)}}
	existingUsers := chatChannel.GetUsers()
	for _, existingUser := range existingUsers {
		var existingUserSession *models.Session
		existingUserSession, ok = models.GetSessionByUserId(existingUser.GetId())
		wss.SendOrStoreMessage(
			existingUserSession,
			"ChannelJoinMessage",
			staticResponse,
		)
	}
}
