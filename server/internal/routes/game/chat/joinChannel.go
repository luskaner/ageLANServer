package chat

import (
	"encoding/json"
	"fmt"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/wss"
	"net/http"
	"strconv"
)

func JoinChannel(w http.ResponseWriter, r *http.Request) {
	// FIXME: Channels might show duplicate users (including the count)
	chatChannelIdStr := r.Form.Get("chatroomID")
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
	sess, _ := middleware.Session(r)
	users := game.Users()
	user, _ := users.GetUserById(sess.GetUserId())
	if chatChannel.HasUser(user) {
		i.JSON(&w, i.A{2, "", 0, i.A{}})
		return
	}
	encodedUsers := user.JoinChatChannel(chatChannel)
	i.JSON(&w, i.A{0, chatChannelIdStr, 0, encodedUsers})
	jsonData, err := json.Marshal(i.A{0, chatChannelIdStr, 0, encodedUsers})
	if err != nil {
		panic(err)
	}
	fmt.Println(r.RemoteAddr, string(jsonData))
	staticResponse := i.A{chatChannelIdStr, i.A{0, user.GetProfileInfo(false)}}
	for _, userId := range users.GetUserIds() {
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
