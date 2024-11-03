package relationship

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

func Relationships(gameTitle string, users *models.MainUsers, user *models.MainUser) i.A {
	profileInfo := users.GetProfileInfo(true, func(u *models.MainUser) bool {
		return u != user && u.GetPresence() > 0
	})
	friends := profileInfo
	lastConnection := profileInfo
	if gameTitle == common.GameAoE3 {
		lastConnection = []i.A{}
	} else {
		friends = []i.A{}
	}
	return i.A{0, friends, i.A{}, i.A{}, i.A{}, lastConnection, i.A{}, i.A{}}
}

func GetRelationships(w http.ResponseWriter, r *http.Request) {
	// As we don't have knowledge of Steam/Xbox friends, nor it is supposed to be many players on the server
	// just return all online users as if they were friends (AoE3) or last connections (AoE2)
	sess, _ := middleware.Session(r)
	game := models.G(r)
	users := game.Users()
	currentUser, _ := users.GetUserById(sess.GetUserId())
	i.JSON(&w, Relationships(game.Title(), users, currentUser))
}
