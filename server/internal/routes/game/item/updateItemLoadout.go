package item

import (
	"net/http"

	mapset "github.com/deckarep/golang-set/v2"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

type updateItemLoadoutRequest struct {
	itemIdLoadoutRequest
	createItemLoadoutRequest
}

func UpdateItemLoadout(w http.ResponseWriter, r *http.Request) {
	var req updateItemLoadoutRequest
	err := i.Bind(r, &req)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	game := models.G(r)
	sess := models.SessionOrPanic(r)
	userId := sess.GetUserId()
	u, ok := game.Users().GetUserById(userId)
	if !ok {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	var itemLoadoutEncoded i.A
	_ = u.GetItemLoadouts().WithReadWrite(func(data models.ItemLoadouts) error {
		if item := data.Get(req.Id); item != nil {
			item.Update(req.Name, req.Type, mapset.NewThreadUnsafeSet(req.ItemOrLocIds...))
			itemLoadoutEncoded = item.Encode(userId)
		}
		return nil
	})
	if itemLoadoutEncoded == nil {
		i.JSON(&w, i.A{2, i.A{}})
	} else {
		i.JSON(&w, i.A{0, itemLoadoutEncoded})
	}
}
