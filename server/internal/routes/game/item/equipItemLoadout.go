package item

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

type itemIdLoadoutRequest struct {
	Id int32 `schema:"id"`
}

func EquipItemLoadout(w http.ResponseWriter, r *http.Request) {
	// FIXME: What's the change after equip?
	var req itemIdLoadoutRequest
	err := i.Bind(r, &req)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	game := models.G(r)
	sess := models.SessionOrPanic(r)
	userId := sess.GetUserId()
	u, ok := game.Users().GetUserById(userId)
	if !ok {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	var itemLoadoutEncoded i.A
	_ = u.GetItemLoadouts().WithReadOnly(func(data models.ItemLoadouts) error {
		itemLoadout := data.Get(req.Id)
		if itemLoadout != nil {
			itemLoadoutEncoded = itemLoadout.Encode(userId)
		}
		return nil
	})
	if itemLoadoutEncoded == nil {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
	} else {
		i.JSON(&w, i.A{0, itemLoadoutEncoded, i.A{}})
	}
}
