package playfab

import (
	"fmt"
	"strconv"
	"time"

	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
)

func Items(blessings []Blessings) (catalogItems map[string]playfab.CatalogItem, inventoryItems []playfab.InventoryItem) {
	catalogItems = make(map[string]playfab.CatalogItem)
	for _, b := range blessings {
		prefix := fmt.Sprintf("Item_Season0_%s_", b.EffectName)
		for _, r := range b.KnownRarities {
			if r > -1 {
				name := prefix + strconv.Itoa(r)
				inventoryItems = append(
					inventoryItems,
					playfab.InventoryItem{
						Id:      name,
						StackId: "default",
						Amount:  1,
						Type:    "catalogItem",
					},
				)
				dateFormatted := time.Date(2024, 5, 2, 3, 34, 0, 0, time.UTC).Format(playfab.Iso8601Layout)
				catalogItems[name] =
					playfab.CatalogItem{
						Id:   name,
						Type: "catalogItem",
						AlternateIds: []playfab.CatalogItemAlternativeId{
							{
								"FriendlyId",
								name,
							},
						},
						FriendlyId:       name,
						Title:            playfab.CatalogItemTitle{NEUTRAL: name},
						CreatorEntity:    playfab.CatalogItemCreatorEntity{Id: "C15F9", Type: "title", TypeString: "title"},
						Platforms:        []any{},
						Tags:             []any{},
						CreationDate:     dateFormatted,
						LastModifiedDate: dateFormatted,
						StartDate:        dateFormatted,
						Contents:         []any{},
						Images:           []any{},
						ItemReferences:   []any{},
						DeepLinks:        []any{},
					}
			}
		}
	}
	return
}
