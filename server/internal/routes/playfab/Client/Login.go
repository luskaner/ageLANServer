package Client

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

type customInfoResultPayload struct {
	UserInventory           []any
	UserDataVersion         int
	UserReadOnlyDataVersion int
	CharacterInventories    []any
}

type entityResponse struct {
	Id         string
	Type       string
	TypeString string
}

type treatmentAssignmentResponse struct {
	Variants  []any
	Variables []any
}

type entityTokenResponse struct {
	EntityToken     string
	TokenExpiration string
	entityResponse  `json:"Entity"`
}

type settingsForUserResponse struct {
	NeedsAttribution bool
	GatherDeviceInfo bool
	GatherFocusInfo  bool
}

type loginResponse struct {
	SessionTicket               string
	PlayFabId                   string
	NewlyCreated                bool
	settingsForUserResponse     `json:"SettingsForUser"`
	LastLoginTime               string
	entityTokenResponse         `json:"EntityToken"`
	treatmentAssignmentResponse `json:"TreatmentAssignment"`
	*customInfoResultPayload    `json:"InfoResultPayload,omitempty"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	sessionTicket := uuid.NewString()
	entityToken := uuid.NewString()
	var key string
	var payload *customInfoResultPayload
	if title := models.G(r).Title(); title == common.GameAoE4 {
		key = sessionTicket
		payload = &customInfoResultPayload{
			UserInventory:        []any{},
			CharacterInventories: []any{},
		}
	} else {
		key = entityToken
	}
	id := playfab.AddSession(key)
	shared.RespondOK(
		&w,
		loginResponse{
			SessionTicket: sessionTicket,
			PlayFabId:     id,
			NewlyCreated:  true,
			settingsForUserResponse: settingsForUserResponse{
				NeedsAttribution: false,
				GatherDeviceInfo: true,
				GatherFocusInfo:  true,
			},
			LastLoginTime: shared.FormatDate(now.AddDate(0, 0, -1)),
			entityTokenResponse: entityTokenResponse{
				EntityToken:     entityToken,
				TokenExpiration: shared.FormatDate(now.AddDate(0, 0, 1)),
				entityResponse: entityResponse{
					Id:         uuid.NewString(),
					Type:       "title_player_account",
					TypeString: "title_player_account",
				},
			},
			treatmentAssignmentResponse: treatmentAssignmentResponse{
				Variants:  []any{},
				Variables: []any{},
			},
			customInfoResultPayload: payload,
		},
	)
}
