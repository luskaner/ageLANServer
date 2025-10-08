package chat

import (
	"net/http"
	"strconv"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

func NotifyLeaveChannel(users *models.MainUsers, user *models.MainUser, chatChannelId int32, gameTitle string, clientLibVersion uint16) {
	staticResponse := i.A{strconv.Itoa(int(chatChannelId)), user.GetProfileInfo(false, gameTitle, clientLibVersion)}
	for userId := range users.GetUserIds() {
		if userId == user.GetId() {
			continue
		}
		existingUserSession, ok := models.GetSessionByUserId(userId)
		if ok {
			wss.SendOrStoreMessage(
				existingUserSession,
				"ChannelLeaveMessage",
				staticResponse,
			)
		}
	}
}

func LeaveChannel(w http.ResponseWriter, r *http.Request) {
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
	sess := middleware.SessionOrPanic(r)
	users := game.Users()
	user, ok := users.GetUserById(sess.GetUserId())
	if !ok || !chatChannel.RemoveUser(user) {
		i.JSON(&w, i.A{2})
		return
	}
	i.JSON(&w, i.A{0})
	NotifyLeaveChannel(users, user, chatChannel.GetId(), game.Title(), sess.GetClientLibVersion())
}
