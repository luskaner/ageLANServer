package login

import (
	"net/http"
	"time"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/chat"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/relationship"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	sess := models.SessionOrPanic(r)
	game := models.G(r)
	users := game.Users()
	advertisements := game.Advertisements()
	u, ok := users.GetUserById(sess.GetUserId())
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	if adv := advertisements.GetUserAdvertisement(u.GetId()); adv != nil {
		advertisements.WithWriteLock(adv.GetId(), func() {
			game.Advertisements().UnsafeRemovePeer(adv.GetId(), u.GetId())
		})
	}
	sessions := game.Sessions()
	for channelId, channel := range game.ChatChannels().Iter() {
		if channel.RemoveUser(u) {
			chat.NotifyLeaveChannel(sessions, users, u, channelId, sess.GetClientLibVersion())
			// AoE3 only takes into account the first notify in a readSession return
			// so delay each message by 100ms so they go in different responses
			// otherwise, it would appear as it left the first channel only
			time.Sleep(100 * time.Millisecond)
		}
	}
	relationship.ChangePresence(sess.GetClientLibVersion(), sessions, users, u, 0)
	if game.Title() == common.GameAoE3 || game.Title() == common.GameAoM {
		profileInfo := u.GetProfileInfo(false, sess.GetClientLibVersion())
		for user := range users.GetUserIds() {
			if user != u.GetId() {
				currentSess, currentOk := sessions.GetByUserId(user)
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
	sessions.Delete(sess.Id())
	i.JSON(&w, i.A{0})
}
