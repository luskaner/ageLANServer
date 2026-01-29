package item

import (
	"net/http"
	"slices"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

type updateItemAttributesRequest struct {
	Keys    i.Json[i.Json[[]string]] `json:"attributeKeys"`
	Values  i.Json[i.Json[[]string]] `json:"attributeValues"`
	ItemIds i.Json[[]int32]          `json:"itemInstance_ids"`
	XpGains i.Json[[]int32]          `json:"xpGains"`
}

func UpdateItemAttributes(w http.ResponseWriter, r *http.Request) {
	var req updateItemAttributesRequest
	err := i.Bind(r, &req)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	minLength := min(len(req.Keys.Data.Data), len(req.Values.Data.Data), len(req.ItemIds.Data), len(req.XpGains.Data))
	maxLength := max(len(req.Keys.Data.Data), len(req.Values.Data.Data), len(req.ItemIds.Data), len(req.XpGains.Data))
	if minLength == 0 || maxLength == 0 || minLength != maxLength {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	game := models.G(r)
	sess := models.SessionOrPanic(r)
	u, ok := game.Users().GetUserById(sess.GetUserId())
	if !ok {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	errorCodes := make([]int, minLength)
	itemsEncoded := make([]i.A, minLength)
	_ = u.GetItems().WithReadWrite(func(data *map[int32]models.Item) error {
		for j, itemId := range req.ItemIds.Data {
			// TODO: Handle XP gains
			var itemIdErrorCodes i.A
			attributeKeys := req.Keys.Data.Data
			if item, exists := (*data)[itemId]; !exists {
				itemIdErrorCodes = i.A{2, slices.Repeat(i.A{2}, len(attributeKeys))}
				itemsEncoded[j] = i.A{}
			} else {
				metadata := item.GetMetadata()
				attributeValues := req.Values.Data.Data
				for k := 0; k < len(attributeKeys); k++ {
					attr := attributeKeys[k]
					value := attributeValues[k]
					metadata.UpdateAttribute(attr, value)
				}
				item.IncrementVersion()
				itemIdErrorCodes[j] = itemIdErrorCodes
				itemsEncoded[j] = item.Encode(sess.GetUserId())
			}
		}
		return nil
	})
	i.JSON(&w, i.A{0, errorCodes, itemsEncoded})
}
