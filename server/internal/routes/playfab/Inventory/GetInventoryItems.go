package Inventory

import (
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

type getInventoryItemsResponse struct {
	Items []any
	ETag  string
}

func GetInventoryItems(w http.ResponseWriter, _ *http.Request) {
	shared.RespondOK(
		&w,
		getInventoryItemsResponse{
			Items: []any{},
			ETag:  `1/MQ=="`,
		},
	)
}
