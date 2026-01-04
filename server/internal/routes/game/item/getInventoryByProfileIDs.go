package item

import (
	"net/http"
	"strconv"

	i "github.com/luskaner/ageLANServer/server/internal"
)

type getInventoryByProfileIDsRequest struct {
	ProfileIDs i.Json[[]int32] `json:"profileIDs"`
}

func GetInventoryByProfileIDs(w http.ResponseWriter, r *http.Request) {
	var req getInventoryByProfileIDsRequest
	err := i.Bind(r, &req)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	initialData := make(i.A, len(req.ProfileIDs.Data))
	finalData := make(i.A, len(req.ProfileIDs.Data))
	finalDataArr := i.A{
		// What this mean?
		i.A{1, 0, 0, 0, 10000, 0, 0, 0, 1},
		i.A{2, 0, 1, 0, 10000, 0, 1, 1, 0},
	}
	for j, profileId := range req.ProfileIDs.Data {
		profileIdStr := strconv.Itoa(int(profileId))
		initialData[j] = i.A{
			profileIdStr,
			// And this?
			i.A{},
		}
		finalData[j] = i.A{
			profileIdStr,
			finalDataArr,
		}
	}
	i.JSON(&w, i.A{0, initialData, finalData})
}
