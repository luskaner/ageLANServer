package Client

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

type punchCardProgress struct {
	Holes                     uint8
	DateOfMostRecentHolePunch playfab.CustomTime
}

type punchCardProgressValue struct {
	playfab.Value[punchCardProgress]
}

type missionProgress struct {
	State                 string
	RewardsAwarded        string
	CompletionCountEasy   uint32
	CompletionCountMedium uint32
	CompletionCountHard   uint32
}

type missionProgressValue struct {
	playfab.Value[missionProgress]
	name string
}

type getUserReadOnlyDataResponse struct {
	DataVersion int32
	Data        map[string]playfab.ValueLike
}

type getUserReadonlyDataRequest struct {
	IfChangedFromDataVersion *uint32
	Keys                     []string
	PlayFabId                *string
}

func getType(key string) string {
	switch {
	case key == "PunchCardProgress":
		return "punchCardProgress"
	case strings.HasPrefix(key, "Mission_"):
		return "missionProgress"
	default:
		return ""
	}
}

func getValue(key string) playfab.ValueLike {
	switch getType(key) {
	case "punchCardProgress":
		return &punchCardProgressValue{
			playfab.Value[punchCardProgress]{
				Val: &punchCardProgress{
					Holes: 0,
					DateOfMostRecentHolePunch: playfab.CustomTime{
						Time:   time.Date(2024, 5, 2, 3, 34, 0, 0, time.UTC),
						Format: time.RFC3339,
					},
				},
			},
		}
	case "missionProgress":
		return &missionProgressValue{
			playfab.Value[missionProgress]{
				Val: &missionProgress{
					State:               "Completed",
					RewardsAwarded:      "Hard",
					CompletionCountHard: 1,
				},
			},
			key,
		}
	default:
		return nil
	}
}

func GetUserReadOnlyData(w http.ResponseWriter, r *http.Request) {
	var req getUserReadonlyDataRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || len(req.Keys) == 0 {
		shared.RespondBadRequest(&w)
		return
	}
	var response getUserReadOnlyDataResponse
	response.Data = make(map[string]playfab.ValueLike)
	for _, key := range req.Keys {
		if val := getValue(key); val != nil {
			if err = val.Prepare(); err != nil {
				shared.RespondBadRequest(&w)
				return
			}
			response.Data[key] = val
		}
	}
	shared.RespondOK(&w, response)
}
