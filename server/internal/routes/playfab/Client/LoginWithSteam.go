package Client

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

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

type loginWithSteamResponse struct {
	SessionTicket               string
	PlayFabId                   string
	NewlyCreated                bool
	settingsForUserResponse     `json:"SettingsForUser"`
	LastLoginTime               string
	entityTokenResponse         `json:"EntityToken"`
	treatmentAssignmentResponse `json:"TreatmentAssignment"`
}

func LoginWithSteam(w http.ResponseWriter, _ *http.Request) {
	now := time.Now().UTC()
	entityToken := uuid.NewString()
	id := playfab.AddSession(entityToken)
	shared.RespondOK(
		&w,
		loginWithSteamResponse{
			SessionTicket: uuid.NewString(),
			PlayFabId:     id,
			NewlyCreated:  true,
			settingsForUserResponse: settingsForUserResponse{
				NeedsAttribution: false,
				GatherDeviceInfo: true,
				GatherFocusInfo:  true,
			},
			LastLoginTime: shared.FormatDate(time.Date(2025, 11, 12, 3, 34, 0, 0, time.UTC)),
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
		},
	)
}
