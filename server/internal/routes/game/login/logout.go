package login

import (
	"github.com/luskaner/aoe2DELanServer/common"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/chat"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/relationship"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/wss"
	"net/http"
	"time"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := middleware.Session(r)
	game := models.G(r)
	users := game.Users()
	u, _ := users.GetUserById(sess.GetUserId())
	if adv := u.GetAdvertisement(); adv != nil {
		game.Advertisements().RemovePeer(adv, u)
	}
	channels := u.GetChannels()
	for j, channel := range channels {
		u.LeaveChatChannel(channel)
		chat.NotifyLeaveChannel(users, u, channel.GetId())
		// AoE3 only takes into account the first notify in a readSession return
		// so delay each message by 100ms so they go in different responses
		// otherwise, it would appear as it left the first channel only
		if j != len(channels)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	relationship.ChangePresence(users, u, 0)
	if game.Title() == common.GameAoE3 {
		profileInfo := u.GetProfileInfo(false)
		for _, user := range users.GetUserIds() {
			if user != u.GetId() {
				currentSess, currentOk := models.GetSessionByUserId(user)
				if currentOk {
					wss.SendOrStoreMessage(
						currentSess,
						"FriendClearMessage",
						i.A{profileInfo, 0},
					)
				}
			}
		}
	}
	sess.Delete()
	i.JSON(&w, i.A{0})
}
