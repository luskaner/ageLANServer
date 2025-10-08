package chat

import (
	"net/http"
	"strconv"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

func JoinChannel(w http.ResponseWriter, r *http.Request) {
	// FIXME: Channels might show duplicate users (including the count)
	chatChannelIdStr := r.FormValue("chatroomID")
	if chatChannelIdStr == "" {
		i.JSON(&w, i.A{2, "", 0, i.A{}})
		return
	}
	chatChannelId, err := strconv.ParseInt(chatChannelIdStr, 10, 32)
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
	sess := middleware.SessionOrPanic(r)
	users := game.Users()
	user, ok := users.GetUserById(sess.GetUserId())
	if !ok {
		i.JSON(&w, i.A{2, "", 0, i.A{}})
		return
	}
	exists, encodedUsers := chatChannel.AddUser(user, game.Title(), sess.GetClientLibVersion())
	if exists {
		i.JSON(&w, i.A{2, "", 0, i.A{}})
		return
	}
	i.JSON(&w, i.A{0, chatChannelIdStr, 0, encodedUsers})
	staticResponse := i.A{chatChannelIdStr, i.A{0, user.GetProfileInfo(false, game.Title(), sess.GetClientLibVersion())}}
	for userId := range users.GetUserIds() {
		var existingUserSession *models.Session
		existingUserSession, ok = models.GetSessionByUserId(userId)
		if ok {
			wss.SendOrStoreMessage(
				existingUserSession,
				"ChannelJoinMessage",
				staticResponse,
			)
		}
	}
}
