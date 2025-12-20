package Catalog

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/athens"
	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

type getItemsRequest struct {
	AlternateIds []any
	CustomTags   any
	Entity       any
	Ids          []string
}

type getItemsResponse struct {
	Items []playfab.CatalogItem
}

func GetItems(w http.ResponseWriter, r *http.Request) {
	var req getItemsRequest
	err := i.Bind(r, &req)
	if err != nil || len(req.Ids) == 0 {
		shared.RespondBadRequest(&w)
		return
	}
	game := models.Gg[*athens.Game](r)
	catalogItemsMap := game.CatalogItems
	var catalogItems []playfab.CatalogItem
	for _, id := range req.Ids {
		catalogItem, ok := catalogItemsMap[id]
		if ok {
			catalogItems = append(catalogItems, catalogItem)
		}
	}
	shared.RespondOK(
		&w,
		getItemsResponse{
			Items: catalogItems,
		},
	)
}
