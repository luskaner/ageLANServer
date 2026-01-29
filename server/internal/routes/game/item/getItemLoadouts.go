package item

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func GetItemLoadouts(w http.ResponseWriter, r *http.Request) {
	var itemsEncoded i.A
	sess := models.SessionOrPanic(r)
	game := models.G(r)
	userId := sess.GetUserId()
	u, _ := game.Users().GetUserById(userId)
	if itemLoadouts := u.GetItemLoadouts(); itemLoadouts != nil {
		_ = itemLoadouts.WithReadOnly(func(data models.ItemLoadouts) error {
			for loadout := range data.Iter() {
				itemsEncoded = append(itemsEncoded, loadout.Encode(userId))
			}
			return nil
		})
	}
	i.JSON(&w, i.A{0, itemsEncoded})
}
