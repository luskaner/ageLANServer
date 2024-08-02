package item

import (
	"encoding/json"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
	"strconv"
)

func GetInventoryByProfileIDs(w http.ResponseWriter, r *http.Request) {
	profileIdsStr := r.URL.Query().Get("profileIDs")
	var profileIds []int32
	err := json.Unmarshal([]byte(profileIdsStr), &profileIds)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	initialData := make(i.A, len(profileIds))
	finalData := make(i.A, len(profileIds))
	finalDataArr := i.A{
		// What this mean?
		i.A{1, 0, 0, 0, 10000, 0, 0, 0, 1},
		i.A{2, 0, 1, 0, 10000, 0, 1, 1, 0},
	}
	for j, profileId := range profileIds {
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
