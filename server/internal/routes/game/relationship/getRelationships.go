package relationship

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func Relationships(gameTitle string, clientLibVersion uint16, users *models.MainUsers, user *models.MainUser) i.A {
	profileInfo := users.GetProfileInfo(true, func(u *models.MainUser) bool {
		return u != user && u.GetPresence() > 0
	}, gameTitle, clientLibVersion)
	friends := profileInfo
	lastConnection := profileInfo
	if gameTitle == common.GameAoE3 || gameTitle == common.GameAoM {
		lastConnection = []i.A{}
	} else {
		friends = []i.A{}
	}
	return i.A{0, friends, i.A{}, i.A{}, i.A{}, lastConnection, i.A{}, i.A{}}
}

func GetRelationships(w http.ResponseWriter, r *http.Request) {
	// As we don't have knowledge of Steam/Xbox friends, nor it is supposed to be many players on the server
	// just return all online users as if they were friends (AoE3/AoM) or last connections (AoE2)
	sess := middleware.Session(r)
	game := models.G(r)
	users := game.Users()
	currentUser, ok := users.GetUserById(sess.GetUserId())
	if ok {
		i.JSON(&w, Relationships(game.Title(), sess.GetClientLibVersion(), users, currentUser))
	} else {
		i.JSON(&w, i.A{0, []i.A{}, i.A{}, i.A{}, i.A{}, []i.A{}, i.A{}, i.A{}})
	}
}
