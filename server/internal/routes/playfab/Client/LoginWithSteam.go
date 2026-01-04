package Client

import (
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/athens"
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

type loginWithSteamRequest struct {
	SteamTicket string
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

func LoginWithSteam(w http.ResponseWriter, r *http.Request) {
	var req loginWithSteamRequest
	err := i.Bind(r, &req)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)
	if err != nil {
		shared.RespondBadRequest(&w)
		return
	}
	var steamId uint64
	steamId, err = playfab.ParseSteamIDHex(req.SteamTicket)
	if err != nil {
		shared.RespondBadRequest(&w)
		return
	}
	now := time.Now().UTC()
	game := models.Gg[*athens.Game](r)
	sessions := game.PlayfabSessions
	id := sessions.Create(game.Users(), steamId)
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
				EntityToken:     id,
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
