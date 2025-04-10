package shared

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func ParseParameters(r *http.Request) (parseError bool, advId int32, length int, profileIds []int32, raceIds []int32, teamIds []int32) {
	profileIdsStr := r.PostFormValue("profile_ids")
	err := json.Unmarshal([]byte(profileIdsStr), &profileIds)
	if err != nil {
		parseError = true
		return
	}
	raceIdsStr := r.PostFormValue("race_ids")
	err = json.Unmarshal([]byte(raceIdsStr), &raceIds)
	if err != nil {
		parseError = true
		return
	}
	teamIdsStr := r.PostFormValue("teamIDs")
	err = json.Unmarshal([]byte(teamIdsStr), &teamIds)
	if err != nil {
		parseError = true
		return
	}
	if min(len(profileIds), len(raceIds), len(teamIds)) != max(len(profileIds), len(raceIds), len(teamIds)) {
		parseError = true
		return
	}
	advIdStr := r.PostFormValue("match_id")
	advIdInt64, err := strconv.ParseInt(advIdStr, 10, 32)
	if err != nil {
		parseError = true
		return
	}
	advId = int32(advIdInt64)
	length = len(profileIds)
	return
}
